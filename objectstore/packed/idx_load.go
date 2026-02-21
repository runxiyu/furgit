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

// packCandidate describes one discovered pack/index pair.
type packCandidate struct {
	// packName is the .pack basename.
	packName string
	// idxName is the .idx basename.
	idxName string
	// mtime is the pack file modification time for initial ordering.
	mtime int64
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

// ensureCandidates discovers pack/index pairs once.
func (store *Store) ensureCandidates() error {
	store.discoverOnce.Do(func() {
		candidates, err := store.discoverCandidates()
		candidateByPack := make(map[string]packCandidate, len(candidates))
		for _, candidate := range candidates {
			candidateByPack[candidate.packName] = candidate
		}
		store.stateMu.Lock()
		store.candidates = candidates
		store.candidateByPack = candidateByPack
		store.discoverErr = err
		store.stateMu.Unlock()
	})

	store.stateMu.RLock()
	err := store.discoverErr
	store.stateMu.RUnlock()
	return err
}

// discoverCandidates scans the objects/pack root and returns sorted pack/index
// pairs.
func (store *Store) discoverCandidates() ([]packCandidate, error) {
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

	candidates := make([]packCandidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".idx") {
			continue
		}

		idxName := entry.Name()
		packName := strings.TrimSuffix(idxName, ".idx") + ".pack"
		packInfo, err := store.root.Stat(packName)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("objectstore/packed: missing pack file for index %q", idxName)
			}
			return nil, err
		}

		candidates = append(candidates, packCandidate{
			packName: packName,
			idxName:  idxName,
			mtime:    packInfo.ModTime().UnixNano(),
		})
	}

	slices.SortFunc(candidates, func(a, b packCandidate) int {
		if a.mtime != b.mtime {
			if a.mtime > b.mtime {
				return -1
			}
			return 1
		}
		return strings.Compare(a.packName, b.packName)
	})

	return candidates, nil
}

// candidateForPack returns one discovered candidate for a pack basename.
func (store *Store) candidateForPack(packName string) (packCandidate, bool) {
	store.stateMu.RLock()
	candidate, ok := store.candidateByPack[packName]
	store.stateMu.RUnlock()
	return candidate, ok
}

// openIndex returns one opened and parsed index, caching it by pack basename.
func (store *Store) openIndex(candidate packCandidate) (*idxFile, error) {
	store.stateMu.RLock()
	if index, ok := store.idxByPack[candidate.packName]; ok {
		store.stateMu.RUnlock()
		return index, nil
	}
	store.stateMu.RUnlock()

	index, err := openIdxFile(store.root, candidate.idxName, candidate.packName, store.algo)
	if err != nil {
		return nil, err
	}

	store.stateMu.Lock()
	if existing, ok := store.idxByPack[candidate.packName]; ok {
		store.stateMu.Unlock()
		_ = index.close()
		return existing, nil
	}
	store.idxByPack[candidate.packName] = index
	store.stateMu.Unlock()
	return index, nil
}

// touchCandidate moves one candidate to the front of the lookup order.
func (store *Store) touchCandidate(packName string) {
	store.stateMu.Lock()
	defer store.stateMu.Unlock()
	for i := range store.candidates {
		if store.candidates[i].packName != packName {
			continue
		}
		if i == 0 {
			return
		}
		candidate := store.candidates[i]
		copy(store.candidates[1:i+1], store.candidates[0:i])
		store.candidates[0] = candidate
		return
	}
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
