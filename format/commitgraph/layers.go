package commitgraph

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"

	"codeberg.org/lindenii/furgit/format/commitgraph/bloom"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// LayerInfo describes one loaded commit-graph layer.
type LayerInfo struct {
	Path      string
	BaseCount uint32
	Commits   uint32
}

type layer struct {
	path       string
	file       *os.File
	data       []byte
	numCommits uint32
	baseCount  uint32
	globalFrom uint32

	chunkOIDFanout    []byte
	chunkOIDLookup    []byte
	chunkCommit       []byte
	chunkGeneration   []byte
	chunkGenerationOv []byte
	chunkExtraEdges   []byte
	chunkBloomIndex   []byte
	chunkBloomData    []byte
	chunkBaseGraphs   []byte

	bloomSettings *bloom.Settings
}

// Layers returns loaded layer metadata in native chain order.
func (reader *Reader) Layers() []LayerInfo {
	out := make([]LayerInfo, 0, len(reader.layers))
	for i := range reader.layers {
		layer := reader.layers[i]
		out = append(out, LayerInfo{
			Path:      layer.path,
			BaseCount: layer.baseCount,
			Commits:   layer.numCommits,
		})
	}

	return out
}

func (reader *Reader) layerByPosition(pos Position) (*layer, error) {
	graphIdx, err := intconv.Uint64ToInt(uint64(pos.Graph))
	if err != nil {
		return nil, err
	}

	if graphIdx < 0 || graphIdx >= len(reader.layers) {
		return nil, &ErrPositionOutOfRange{Pos: pos}
	}

	layer := &reader.layers[graphIdx]
	if pos.Index >= layer.numCommits {
		return nil, &ErrPositionOutOfRange{Pos: pos}
	}

	return layer, nil
}

func layerLookup(layer *layer, oid objectid.ObjectID) (uint32, bool) {
	hashSize := oid.Size()
	first := int(oid.RawBytes()[0])

	var lo uint32
	if first > 0 {
		lo = binary.BigEndian.Uint32(layer.chunkOIDFanout[(first-1)*4 : first*4])
	}

	hi := binary.BigEndian.Uint32(layer.chunkOIDFanout[first*4 : (first+1)*4])
	if hi == 0 || lo >= hi {
		return 0, false
	}

	target := oid.RawBytes()
	left := int(lo)

	right := int(hi) - 1
	for left <= right {
		mid := left + (right-left)/2
		start := mid * hashSize
		end := start + hashSize

		current := layer.chunkOIDLookup[start:end]

		cmp := bytes.Compare(current, target)
		switch {
		case cmp == 0:
			pos, err := intconv.IntToUint32(mid)
			if err != nil {
				return 0, false
			}

			return pos, true
		case cmp < 0:
			left = mid + 1
		default:
			right = mid - 1
		}
	}

	return 0, false
}

func openLayer(root *os.Root, relPath string, algo objectid.Algorithm) (*layer, error) {
	file, err := root.Open(relPath)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	size := info.Size()
	if size < int64(headerSize+fanoutSize+algo.Size()) {
		_ = file.Close()

		return nil, &ErrMalformed{Path: relPath, Reason: "file too short"}
	}

	mapLen, err := intconv.Int64ToUint64(size)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	mapLenInt, err := intconv.Uint64ToInt(mapLen)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	fd, err := intconv.UintptrToInt(file.Fd())
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	data, err := syscall.Mmap(fd, 0, mapLenInt, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	out := &layer{
		path: relPath,
		file: file,
		data: data,
	}

	parseErr := parseLayer(out, algo)
	if parseErr != nil {
		_ = out.close()

		return nil, parseErr
	}

	verifyErr := verifyTrailerHash(out.data, algo, relPath)
	if verifyErr != nil {
		_ = out.close()

		return nil, verifyErr
	}

	return out, nil
}

func parseLayer(layer *layer, algo objectid.Algorithm) error { //nolint:maintidx
	if len(layer.data) < headerSize {
		return &ErrMalformed{Path: layer.path, Reason: "file too short"}
	}

	header := layer.data[:headerSize]

	signature := binary.BigEndian.Uint32(header[:4])
	if signature != fileSignature {
		return &ErrMalformed{Path: layer.path, Reason: "invalid signature"}
	}

	version := header[4]
	if version != fileVersion {
		return &ErrUnsupportedVersion{Version: version}
	}

	expectedHashVersion, err := intconv.Uint32ToUint8(algo.PackHashID())
	if err != nil {
		return err
	}

	hashVersion := header[5]
	if hashVersion != expectedHashVersion {
		return &ErrMalformed{Path: layer.path, Reason: "hash version does not match object format"}
	}

	numChunks := int(header[6])
	baseCount := uint32(header[7])

	tocLen := (numChunks + 1) * chunkEntrySize
	tocStart := headerSize

	tocEnd := tocStart + tocLen
	if tocEnd > len(layer.data) {
		return &ErrMalformed{Path: layer.path, Reason: "truncated chunk table"}
	}

	type tocEntry struct {
		id     uint32
		offset uint64
	}

	entries := make([]tocEntry, 0, numChunks+1)
	for i := range numChunks + 1 {
		entryOff := tocStart + i*chunkEntrySize
		entryData := layer.data[entryOff : entryOff+chunkEntrySize]

		entry := tocEntry{
			id:     binary.BigEndian.Uint32(entryData[:4]),
			offset: binary.BigEndian.Uint64(entryData[4:]),
		}
		entries = append(entries, entry)
	}

	if entries[len(entries)-1].id != 0 {
		return &ErrMalformed{Path: layer.path, Reason: "missing chunk table terminator"}
	}

	trailerStart := len(layer.data) - algo.Size()

	chunks := make(map[uint32][]byte, numChunks)
	for i := range numChunks {
		entry := entries[i]
		if entry.id == 0 {
			return &ErrMalformed{Path: layer.path, Reason: "early chunk table terminator"}
		}

		next := entries[i+1]

		start, err := intconv.Uint64ToInt(entry.offset)
		if err != nil {
			return err
		}

		end, err := intconv.Uint64ToInt(next.offset)
		if err != nil {
			return err
		}

		if start < tocEnd || end < start || end > trailerStart {
			return &ErrMalformed{Path: layer.path, Reason: "invalid chunk offsets"}
		}

		if _, exists := chunks[entry.id]; exists {
			return &ErrMalformed{Path: layer.path, Reason: "duplicate chunk id"}
		}

		chunks[entry.id] = layer.data[start:end]
	}

	oidf := chunks[chunkOIDF]
	if len(oidf) != fanoutSize {
		return &ErrMalformed{Path: layer.path, Reason: "invalid OIDF length"}
	}

	layer.chunkOIDFanout = oidf
	layer.numCommits = binary.BigEndian.Uint32(oidf[fanoutSize-4:])

	for i := range 255 {
		cur := binary.BigEndian.Uint32(oidf[i*4 : (i+1)*4])

		next := binary.BigEndian.Uint32(oidf[(i+1)*4 : (i+2)*4])
		if cur > next {
			return &ErrMalformed{Path: layer.path, Reason: "non-monotonic OIDF fanout"}
		}
	}

	hashSize := algo.Size()

	hashSizeU64, err := intconv.IntToUint64(hashSize)
	if err != nil {
		return err
	}

	oidl := chunks[chunkOIDL]
	oidlWantLen64 := uint64(layer.numCommits) * hashSizeU64

	oidlWantLen, err := intconv.Uint64ToInt(oidlWantLen64)
	if err != nil {
		return err
	}

	if len(oidl) != oidlWantLen {
		return &ErrMalformed{Path: layer.path, Reason: "invalid OIDL length"}
	}

	layer.chunkOIDLookup = oidl

	stride := hashSize + 16

	strideU64, err := intconv.IntToUint64(stride)
	if err != nil {
		return err
	}

	cdat := chunks[chunkCDAT]
	cdatWantLen64 := uint64(layer.numCommits) * strideU64

	cdatWantLen, err := intconv.Uint64ToInt(cdatWantLen64)
	if err != nil {
		return err
	}

	if len(cdat) != cdatWantLen {
		return &ErrMalformed{Path: layer.path, Reason: "invalid CDAT length"}
	}

	layer.chunkCommit = cdat

	gda2 := chunks[chunkGDA2]
	if len(gda2) != 0 {
		wantLen64 := uint64(layer.numCommits) * 4

		wantLen, err := intconv.Uint64ToInt(wantLen64)
		if err != nil {
			return err
		}

		if len(gda2) != wantLen {
			return &ErrMalformed{Path: layer.path, Reason: "invalid GDA2 length"}
		}

		layer.chunkGeneration = gda2
	}

	gdo2 := chunks[chunkGDO2]
	if len(gdo2) != 0 {
		if len(gdo2)%8 != 0 {
			return &ErrMalformed{Path: layer.path, Reason: "invalid GDO2 length"}
		}

		layer.chunkGenerationOv = gdo2
	}

	edge := chunks[chunkEDGE]
	if len(edge) != 0 {
		if len(edge)%4 != 0 {
			return &ErrMalformed{Path: layer.path, Reason: "invalid EDGE length"}
		}

		layer.chunkExtraEdges = edge
	}

	base := chunks[chunkBASE]
	if baseCount == 0 {
		if len(base) != 0 {
			return &ErrMalformed{Path: layer.path, Reason: "unexpected BASE chunk"}
		}
	} else {
		wantLen64 := uint64(baseCount) * hashSizeU64

		wantLen, err := intconv.Uint64ToInt(wantLen64)
		if err != nil {
			return err
		}

		if len(base) != wantLen {
			return &ErrMalformed{Path: layer.path, Reason: "invalid BASE length"}
		}

		layer.chunkBaseGraphs = base
	}

	layer.baseCount = baseCount

	bidx := chunks[chunkBIDX]

	bdat := chunks[chunkBDAT]
	if len(bidx) != 0 || len(bdat) != 0 { //nolint:nestif
		if len(bidx) == 0 || len(bdat) == 0 {
			return &ErrMalformed{Path: layer.path, Reason: "BIDX/BDAT must both be present"}
		}

		bidxWantLen64 := uint64(layer.numCommits) * 4

		bidxWantLen, err := intconv.Uint64ToInt(bidxWantLen64)
		if err != nil {
			return err
		}

		if len(bidx) != bidxWantLen {
			return &ErrMalformed{Path: layer.path, Reason: "invalid BIDX length"}
		}

		if len(bdat) < bloom.DataHeaderSize {
			return &ErrMalformed{Path: layer.path, Reason: "invalid BDAT length"}
		}

		settings, err := bloom.ParseSettings(bdat)
		if err != nil {
			return err
		}

		prev := uint32(0)

		for i := range layer.numCommits {
			off := int(i) * 4

			cur := binary.BigEndian.Uint32(bidx[off : off+4])
			if i > 0 && cur < prev {
				return &ErrMalformed{Path: layer.path, Reason: "non-monotonic BIDX"}
			}

			bdatDataLen := len(bdat) - bloom.DataHeaderSize

			bdatDataLenU32, err := intconv.IntToUint32(bdatDataLen)
			if err != nil {
				return err
			}

			if cur > bdatDataLenU32 {
				return &ErrMalformed{Path: layer.path, Reason: "BIDX offset out of range"}
			}

			prev = cur
		}

		layer.chunkBloomIndex = bidx
		layer.chunkBloomData = bdat
		layer.bloomSettings = settings
	}

	return nil
}

func closeLayers(layers []layer) {
	for i := len(layers) - 1; i >= 0; i-- {
		_ = layers[i].close()
	}
}

func (layer *layer) close() error {
	var closeErr error

	if layer.data != nil {
		err := syscall.Munmap(layer.data)
		if err != nil {
			closeErr = err
		}

		layer.data = nil
	}

	if layer.file != nil {
		err := layer.file.Close()
		if err != nil && closeErr == nil {
			closeErr = err
		}

		layer.file = nil
	}

	return closeErr
}
