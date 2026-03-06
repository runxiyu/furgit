package read

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/format/commitgraph"
	"codeberg.org/lindenii/furgit/format/commitgraph/bloom"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

func parseLayer(layer *layer, algo objectid.Algorithm) error { //nolint:maintidx
	if len(layer.data) < commitgraph.HeaderSize {
		return &ErrMalformed{Path: layer.path, Reason: "file too short"}
	}

	header := layer.data[:commitgraph.HeaderSize]

	signature := binary.BigEndian.Uint32(header[:4])
	if signature != commitgraph.FileSignature {
		return &ErrMalformed{Path: layer.path, Reason: "invalid signature"}
	}

	version := header[4]
	if version != commitgraph.FileVersion {
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

	tocLen := (numChunks + 1) * commitgraph.ChunkEntrySize
	tocStart := commitgraph.HeaderSize

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
		entryOff := tocStart + i*commitgraph.ChunkEntrySize
		entryData := layer.data[entryOff : entryOff+commitgraph.ChunkEntrySize]

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

	oidf := chunks[commitgraph.ChunkOIDF]
	if len(oidf) != commitgraph.FanoutSize {
		return &ErrMalformed{Path: layer.path, Reason: "invalid OIDF length"}
	}

	layer.chunkOIDFanout = oidf
	layer.numCommits = binary.BigEndian.Uint32(oidf[commitgraph.FanoutSize-4:])

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

	oidl := chunks[commitgraph.ChunkOIDL]
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

	cdat := chunks[commitgraph.ChunkCDAT]
	cdatWantLen64 := uint64(layer.numCommits) * strideU64

	cdatWantLen, err := intconv.Uint64ToInt(cdatWantLen64)
	if err != nil {
		return err
	}

	if len(cdat) != cdatWantLen {
		return &ErrMalformed{Path: layer.path, Reason: "invalid CDAT length"}
	}

	layer.chunkCommit = cdat

	gda2 := chunks[commitgraph.ChunkGDA2]
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

	gdo2 := chunks[commitgraph.ChunkGDO2]
	if len(gdo2) != 0 {
		if len(gdo2)%8 != 0 {
			return &ErrMalformed{Path: layer.path, Reason: "invalid GDO2 length"}
		}

		layer.chunkGenerationOv = gdo2
	}

	edge := chunks[commitgraph.ChunkEDGE]
	if len(edge) != 0 {
		if len(edge)%4 != 0 {
			return &ErrMalformed{Path: layer.path, Reason: "invalid EDGE length"}
		}

		layer.chunkExtraEdges = edge
	}

	base := chunks[commitgraph.ChunkBASE]
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

	bidx := chunks[commitgraph.ChunkBIDX]

	bdat := chunks[commitgraph.ChunkBDAT]
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
