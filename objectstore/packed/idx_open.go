package packed

import (
	"fmt"
	"os"
	"syscall"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

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

// candidateForPack returns one discovered candidate for a pack basename.
func (store *Store) candidateForPack(packName string) (packCandidate, bool) {
	store.candidatesMu.RLock()
	candidate, ok := store.candidateByPack[packName]
	store.candidatesMu.RUnlock()
	return candidate, ok
}

// openIndex returns one opened and parsed index, caching it by pack basename.
func (store *Store) openIndex(candidate packCandidate) (*idxFile, error) {
	store.idxMu.RLock()
	if index, ok := store.idxByPack[candidate.packName]; ok {
		store.idxMu.RUnlock()
		return index, nil
	}
	store.idxMu.RUnlock()

	index, err := openIdxFile(store.root, candidate.idxName, candidate.packName, store.algo)
	if err != nil {
		return nil, err
	}

	store.idxMu.Lock()
	if existing, ok := store.idxByPack[candidate.packName]; ok {
		store.idxMu.Unlock()
		_ = index.close()
		return existing, nil
	}
	store.idxByPack[candidate.packName] = index
	store.idxMu.Unlock()
	return index, nil
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
