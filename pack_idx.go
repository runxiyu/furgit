package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

const (
	idxMagic    = 0xff744f63
	idxVersion2 = 2
)

type packIndex struct {
	repo     *Repository
	idxRel   string
	packPath string

	loadOnce sync.Once
	loadErr  error

	numObjects int
	fanout     []byte
	names      []byte
	crcs       []byte
	offset32   []byte
	offset64   []byte
	data       []byte

	closeOnce sync.Once
}

func (pi *packIndex) Close() error {
	if pi == nil {
		return nil
	}
	var closeErr error
	pi.closeOnce.Do(func() {
		if len(pi.data) > 0 {
			if err := syscall.Munmap(pi.data); closeErr == nil {
				closeErr = err
			}
			pi.data = nil
			pi.fanout = nil
			pi.names = nil
			pi.crcs = nil
			pi.offset32 = nil
			pi.offset64 = nil
			pi.numObjects = 0
		}
	})
	return closeErr
}

func (pi *packIndex) ensureLoaded() error {
	pi.loadOnce.Do(func() {
		pi.loadErr = pi.load()
	})
	return pi.loadErr
}

func (pi *packIndex) load() error {
	if pi.repo == nil {
		return ErrInvalidObject
	}
	f, err := os.Open(pi.repo.repoPath(pi.idxRel))
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	if stat.Size() < 8+256*4 {
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
	err = pi.parse(region)
	if err != nil {
		_ = syscall.Munmap(region)
		return err
	}
	pi.data = region
	return nil
}

func (repo *Repository) packIndexes() ([]*packIndex, error) {
	repo.packIdxOnce.Do(func() {
		repo.packIdx, repo.packIdxErr = repo.loadPackIndexes()
	})
	return repo.packIdx, repo.packIdxErr
}

func (repo *Repository) loadPackIndexes() ([]*packIndex, error) {
	dir := filepath.Join(repo.rootPath, "objects", "pack")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	idxs := make([]*packIndex, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".idx") {
			continue
		}
		rel := filepath.Join("objects", "pack", entry.Name())
		packRel := strings.TrimSuffix(rel, ".idx") + ".pack"
		idxs = append(idxs, &packIndex{
			repo:     repo,
			idxRel:   rel,
			packPath: packRel,
		})
	}
	if len(idxs) == 0 {
		return nil, ErrNotFound
	}
	return idxs, nil
}

func (pi *packIndex) parse(buf []byte) error {
	if len(buf) < 8+256*4 {
		return ErrInvalidObject
	}
	if readBE32(buf[0:4]) != idxMagic {
		return ErrInvalidObject
	}
	if readBE32(buf[4:8]) != idxVersion2 {
		return ErrInvalidObject
	}

	const fanoutBytes = 256 * 4
	fanoutStart := 8
	fanoutEnd := fanoutStart + fanoutBytes
	if fanoutEnd > len(buf) {
		return ErrInvalidObject
	}
	pi.fanout = buf[fanoutStart:fanoutEnd]
	nobj := int(readBE32(pi.fanout[len(pi.fanout)-4:]))

	namesStart := fanoutEnd
	namesEnd := namesStart + nobj*pi.repo.hashSize
	if namesEnd > len(buf) {
		return ErrInvalidObject
	}

	crcStart := namesEnd
	crcEnd := crcStart + nobj*4
	if crcEnd > len(buf) {
		return ErrInvalidObject
	}

	off32Start := crcEnd
	off32End := off32Start + nobj*4
	if off32End > len(buf) {
		return ErrInvalidObject
	}

	pi.offset32 = buf[off32Start:off32End]

	off64Start := off32End
	trailerStart := len(buf) - 2*pi.repo.hashSize
	if trailerStart < off64Start {
		return ErrInvalidObject
	}
	if (trailerStart-off64Start)%8 != 0 {
		return ErrInvalidObject
	}
	off64End := trailerStart
	pi.offset64 = buf[off64Start:off64End]

	pi.numObjects = nobj
	pi.names = buf[namesStart:namesEnd]
	pi.crcs = buf[crcStart:crcEnd]
	return nil
}

func readBE32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func readBE64(b []byte) uint64 {
	_ = b[7]
	return (uint64(b[0]) << 56) | (uint64(b[1]) << 48) |
		(uint64(b[2]) << 40) | (uint64(b[3]) << 32) |
		(uint64(b[4]) << 24) | (uint64(b[5]) << 16) |
		(uint64(b[6]) << 8) | uint64(b[7])
}

func (pi *packIndex) fanoutEntry(i int) uint32 {
	if len(pi.fanout) == 0 {
		return 0
	}
	entries := len(pi.fanout) / 4
	if i < 0 || i >= entries {
		return 0
	}
	start := i * 4
	return readBE32(pi.fanout[start : start+4])
}

func (pi *packIndex) offset(idx int) (uint64, error) {
	start := idx * 4
	word := readBE32(pi.offset32[start : start+4])
	if word&0x80000000 == 0 {
		return uint64(word), nil
	}
	pos := int(word & 0x7fffffff)
	entries := len(pi.offset64) / 8
	if pos < 0 || pos >= entries {
		return 0, errors.New("furgit: pack: corrupt 64-bit offset table")
	}
	base := pos * 8
	return readBE64(pi.offset64[base : base+8]), nil
}

func (pi *packIndex) lookup(id Hash) (packlocation, error) {
	err := pi.ensureLoaded()
	if err != nil {
		return packlocation{}, err
	}
	if id.size != pi.repo.hashSize {
		return packlocation{}, fmt.Errorf("furgit: hash size mismatch: got %d, expected %d", id.size, pi.repo.hashSize)
	}
	first := int(id.data[0])
	var lo int
	if first > 0 {
		lo = int(pi.fanoutEntry(first - 1))
	}
	hi := int(pi.fanoutEntry(first))
	idx, found := bsearchHash(pi.names, pi.repo.hashSize, lo, hi, id)
	if !found {
		return packlocation{}, ErrNotFound
	}
	ofs, err := pi.offset(idx)
	if err != nil {
		return packlocation{}, err
	}
	return packlocation{
		PackPath: pi.packPath,
		Offset:   ofs,
	}, nil
}

func bsearchHash(names []byte, stride, lo, hi int, want Hash) (int, bool) {
	for lo < hi {
		mid := lo + (hi-lo)/2
		cmp := compareHash(names, stride, mid, want.data[:stride])
		if cmp == 0 {
			return mid, true
		}
		if cmp > 0 {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo, false
}

func compareHash(names []byte, stride, idx int, want []byte) int {
	base := idx * stride
	end := base + stride
	return bytes.Compare(names[base:end], want)
}
