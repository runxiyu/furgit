package packed

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"syscall"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// location identifies one object entry in a specific pack file.
type location struct {
	packName string
	offset   uint64
}

// idxFile stores one mapped and validated idx v2 file.
type idxFile struct {
	// idxName is the basename of this .idx file.
	idxName string
	// packName is the matching .pack basename.
	packName string
	// algo is the hash algorithm encoded by the index.
	algo objectid.Algorithm

	// file is the opened index file descriptor.
	file *os.File
	// data is the mapped index bytes.
	data []byte

	// fanout stores fanout table values.
	fanout [256]uint32
	// numObjects equals fanout[255].
	numObjects int

	// namesOffset starts the sorted object-id table.
	namesOffset int
	// offset32Offset starts the 32-bit offset table.
	offset32Offset int
	// offset64Offset starts the 64-bit offset table.
	offset64Offset int
	// offset64Count is the number of 64-bit offset entries.
	offset64Count int
}

// loadIndexes loads and validates all .idx files under objects/pack.
func (store *Store) loadIndexes() ([]*idxFile, error) {
	dir, err := store.root.Open(".")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = dir.Close() }()
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	idxNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".idx") {
			continue
		}
		idxNames = append(idxNames, entry.Name())
	}
	slices.Sort(idxNames)

	out := make([]*idxFile, 0, len(idxNames))
	for _, idxName := range idxNames {
		packName := strings.TrimSuffix(idxName, ".idx") + ".pack"
		if _, err := store.root.Stat(packName); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("objectstore/packed: missing pack file for index %q", idxName)
			}
			return nil, err
		}
		index, err := openIdxFile(store.root, idxName, packName, store.algo)
		if err != nil {
			for _, loaded := range out {
				_ = loaded.close()
			}
			return nil, err
		}
		out = append(out, index)
	}
	return out, nil
}

// openIdxFile maps and validates one idx v2 file.
func openIdxFile(root *os.Root, idxName, packName string, algo objectid.Algorithm) (*idxFile, error) {
	file, err := root.Open(idxName)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	size := info.Size()
	if size < 0 || size > int64(int(^uint(0)>>1)) {
		_ = file.Close()
		return nil, fmt.Errorf("objectstore/packed: idx %q has unsupported size", idxName)
	}
	fd, err := intconv.UintptrToInt(file.Fd())
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	data, err := syscall.Mmap(fd, 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	index := &idxFile{
		idxName:  idxName,
		packName: packName,
		algo:     algo,
		file:     file,
		data:     data,
	}
	if err := index.parse(); err != nil {
		_ = index.close()
		return nil, err
	}
	return index, nil
}

// close unmaps and closes one idx handle.
func (index *idxFile) close() error {
	var closeErr error
	if index.data != nil {
		if err := syscall.Munmap(index.data); err != nil && closeErr == nil {
			closeErr = err
		}
		index.data = nil
	}
	if index.file != nil {
		if err := index.file.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		index.file = nil
	}
	return closeErr
}
