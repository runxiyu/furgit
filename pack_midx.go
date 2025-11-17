package furgit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

const (
	midxMagic   = 0x4d494458 // MIDX
	midxVersion = 1

	midxOIDVersionSHA1   = 1
	midxOIDVersionSHA256 = 2

	chunkPNAM = 0x504e414d // PNAM
	chunkOIDF = 0x4f494446 // OIDF
	chunkOIDL = 0x4f49444c // OIDL
	chunkOOFF = 0x4f4f4646 // OOFF
	chunkLOFF = 0x4c4f4646 // LOFF
)

type multiPackIndex struct {
	repo *Repository

	loadOnce sync.Once
	loadErr  error

	numPacks   int
	numObjects int
	packNames  []string
	fanout     []byte
	oids       []byte
	offsets    []byte
	largeOffs  []byte
	data       []byte

	closeOnce sync.Once
}

func (midx *multiPackIndex) Close() error {
	if midx == nil {
		return nil
	}
	var closeErr error
	midx.closeOnce.Do(func() {
		if len(midx.data) > 0 {
			if err := syscall.Munmap(midx.data); closeErr == nil {
				closeErr = err
			}
			midx.data = nil
			midx.fanout = nil
			midx.oids = nil
			midx.offsets = nil
			midx.largeOffs = nil
			midx.packNames = nil
			midx.numObjects = 0
			midx.numPacks = 0
		}
	})
	return closeErr
}

func (midx *multiPackIndex) ensureLoaded() error {
	midx.loadOnce.Do(func() {
		midx.loadErr = midx.load()
	})
	return midx.loadErr
}

func (midx *multiPackIndex) load() error {
	if midx.repo == nil {
		return ErrInvalidObject
	}
	path := midx.repo.repoPath(filepath.Join("objects", "pack", "multi-pack-index"))
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	if stat.Size() < 12 {
		_ = f.Close()
		return ErrInvalidObject
	}
	region, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(stat.Size()),
		syscall.PROT_READ,
		syscall.MAP_PRIVATE,
	)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		_ = syscall.Munmap(region)
		return err
	}
	err = midx.parse(region)
	if err != nil {
		_ = syscall.Munmap(region)
		return err
	}
	midx.data = region
	return nil
}

func (midx *multiPackIndex) parse(buf []byte) error {
	if len(buf) < 12 {
		return ErrInvalidObject
	}

	if readBE32(buf[0:4]) != midxMagic {
		return ErrInvalidObject
	}
	if buf[4] != midxVersion {
		return ErrInvalidObject
	}
	oidVersion := buf[5]
	if oidVersion != midxOIDVersionSHA1 && oidVersion != midxOIDVersionSHA256 {
		return ErrInvalidObject
	}
	numChunks := int(buf[6])
	numBaseMIDX := int(buf[7])
	if numBaseMIDX != 0 {
		return ErrInvalidObject
	}
	numPacks := int(readBE32(buf[8:12]))

	chunkTableStart := 12
	chunkTableSize := (numChunks + 1) * 12
	if len(buf) < chunkTableStart+chunkTableSize {
		return ErrInvalidObject
	}

	chunks := make(map[uint32]int64)
	for i := 0; i < numChunks; i++ {
		chunkStart := chunkTableStart + i*12
		chunkID := readBE32(buf[chunkStart : chunkStart+4])
		chunkOffset := int64(readBE64(buf[chunkStart+4 : chunkStart+12]))
		chunks[chunkID] = chunkOffset
	}

	pnamOffset, ok := chunks[chunkPNAM]
	if !ok {
		return ErrInvalidObject
	}
	if pnamOffset < 0 || pnamOffset >= int64(len(buf)) {
		return ErrInvalidObject
	}

	nextOffset := int64(len(buf))
	for _, offset := range chunks {
		if offset > pnamOffset && offset < nextOffset {
			nextOffset = offset
		}
	}

	pnamData := buf[pnamOffset:nextOffset]
	packNames := make([]string, 0, numPacks)
	start := 0
	for i := 0; i < numPacks; i++ {
		end := start
		for end < len(pnamData) && pnamData[end] != 0 {
			end++
		}
		if end >= len(pnamData) {
			return ErrInvalidObject
		}
		name := string(pnamData[start:end])
		if strings.HasSuffix(name, ".idx") { // why...
			name = name[:len(name)-4] + ".pack"
		}
		packNames = append(packNames, name)
		start = end + 1
	}

	oidfOffset, ok := chunks[chunkOIDF]
	if !ok {
		return ErrInvalidObject
	}
	if oidfOffset < 0 || oidfOffset+256*4 > int64(len(buf)) {
		return ErrInvalidObject
	}
	fanout := buf[oidfOffset : oidfOffset+256*4]
	numObjects := int(readBE32(fanout[len(fanout)-4:]))

	oidlOffset, ok := chunks[chunkOIDL]
	if !ok {
		return ErrInvalidObject
	}
	oidlSize := int64(numObjects) * int64(midx.repo.hashSize)
	if oidlOffset < 0 || oidlOffset+oidlSize > int64(len(buf)) {
		return ErrInvalidObject
	}
	oids := buf[oidlOffset : oidlOffset+oidlSize]

	ooffOffset, ok := chunks[chunkOOFF]
	if !ok {
		return ErrInvalidObject
	}
	ooffSize := int64(numObjects) * 8
	if ooffOffset < 0 || ooffOffset+ooffSize > int64(len(buf)) {
		return ErrInvalidObject
	}
	offsets := buf[ooffOffset : ooffOffset+ooffSize]

	var largeOffs []byte
	if loffOffset, ok := chunks[chunkLOFF]; ok {
		loffEnd := int64(len(buf))
		for _, offset := range chunks {
			if offset > loffOffset && offset < loffEnd {
				loffEnd = offset
			}
		}
		if loffOffset < 0 || loffOffset > int64(len(buf)) {
			return ErrInvalidObject
		}
		largeOffs = buf[loffOffset:loffEnd]
	}

	midx.numPacks = numPacks
	midx.numObjects = numObjects
	midx.packNames = packNames
	midx.fanout = fanout
	midx.oids = oids
	midx.offsets = offsets
	midx.largeOffs = largeOffs
	return nil
}

func (midx *multiPackIndex) lookup(id Hash) (packlocation, error) {
	if len(midx.data) == 0 {
		err := midx.ensureLoaded()
		if err != nil {
			return packlocation{}, err
		}
	}

	if id.size != midx.repo.hashSize {
		return packlocation{}, fmt.Errorf("furgit: hash size mismatch: got %d, expected %d", id.size, midx.repo.hashSize)
	}

	first := int(id.data[0])
	var lo int
	if first > 0 {
		lo = int(readBE32(midx.fanout[(first-1)*4 : first*4]))
	}
	hi := int(readBE32(midx.fanout[first*4 : (first+1)*4]))

	idx, found := bsearchHash(midx.oids, midx.repo.hashSize, lo, hi, id)
	if !found {
		return packlocation{}, ErrNotFound
	}

	offsetEntry := midx.offsets[idx*8 : (idx+1)*8]
	packIntID := readBE32(offsetEntry[0:4])
	offset := readBE32(offsetEntry[4:8])

	if int(packIntID) >= len(midx.packNames) {
		return packlocation{}, ErrInvalidObject
	}

	var finalOffset uint64
	if offset&0x80000000 != 0 {
		if len(midx.largeOffs) == 0 {
			return packlocation{}, ErrInvalidObject
		}
		largeIdx := int(offset & 0x7fffffff)
		if largeIdx*8+8 > len(midx.largeOffs) {
			return packlocation{}, ErrInvalidObject
		}
		finalOffset = readBE64(midx.largeOffs[largeIdx*8 : largeIdx*8+8])
	} else {
		finalOffset = uint64(offset)
	}

	packName := midx.packNames[packIntID]
	packPath := filepath.Join("objects", "pack", packName)

	return packlocation{
		PackPath: packPath,
		Offset:   finalOffset,
	}, nil
}

func (repo *Repository) multiPackIndex() (*multiPackIndex, error) {
	repo.midxOnce.Do(func() {
		repo.midx, repo.midxErr = repo.loadMultiPackIndex()
	})
	return repo.midx, repo.midxErr
}

func (repo *Repository) loadMultiPackIndex() (*multiPackIndex, error) {
	midx := &multiPackIndex{repo: repo}
	err := midx.ensureLoaded()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return midx, nil
}
