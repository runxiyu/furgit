package reading

import (
	"fmt"
	"os"
	"syscall"

	"lindenii.org/go/furgit/internal/intconv"
	objectid "lindenii.org/go/furgit/object/id"
)

// openIndex returns one opened and parsed index, caching it by pack basename.
func (store *Store) openIndex(candidate packCandidate) (*idxFile, error) {
	store.idxMu.RLock()

	index, ok := store.idxByPack[candidate.packName]
	if ok {
		store.idxMu.RUnlock()

		return index, nil
	}

	store.idxMu.RUnlock()

	index, err := openIdxFile(store.root, candidate.idxName, candidate.packName, store.algo)
	if err != nil {
		return nil, err
	}

	store.idxMu.Lock()

	existing, ok := store.idxByPack[candidate.packName]
	if ok {
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

	err = index.parse()
	if err != nil {
		_ = index.close()

		return nil, err
	}

	return index, nil
}
