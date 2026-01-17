package furgit

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"git.sr.ht/~runxiyu/furgit/config"
)

// Repository represents a Git repository.
//
// It is safe to access the same Repository from multiple goroutines
// without additional synchronization.
//
// Objects derived from a Repository must not be used after the Repository
// has been closed.
type Repository struct {
	rootPath string
	hashAlgo hashAlgorithm
	hashSize int

	packIdxOnce sync.Once
	packIdx     []*packIndex
	packIdxErr  error

	packFiles   map[string]*packFile
	packFilesMu sync.RWMutex
	closeOnce   sync.Once
}

// OpenRepository opens the repository at the provided path.
//
// The path is expected to be the actual repository directory, i.e.,
// the repository itself for bare repositories, or the .git
// subdirectory for non-bare repositories.
func OpenRepository(path string) (*Repository, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrInvalidObject
	}

	cfgPath := filepath.Join(path, "config")
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("furgit: unable to open config: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	cfg, err := config.ParseConfig(f)
	if err != nil {
		return nil, fmt.Errorf("furgit: failed to parse config: %w", err)
	}

	algo := cfg.Get("extensions", "", "objectformat")
	if algo == "" {
		algo = "sha1"
	}

	var hashAlgo hashAlgorithm
	switch algo {
	case "sha1":
		hashAlgo = hashAlgoSHA1
	case "sha256":
		hashAlgo = hashAlgoSHA256
	default:
		return nil, fmt.Errorf("furgit: unsupported hash algorithm %q", algo)
	}

	hashSize := hashAlgo.size()
	if hashSize == 0 {
		return nil, fmt.Errorf("furgit: unsupported hash algorithm %q", algo)
	}
	if _, ok := hashFuncs[hashAlgo]; !ok {
		return nil, fmt.Errorf("furgit: hash algorithm %q is not supported by the hash functions provided by this build", algo)
	}

	return &Repository{
		rootPath:  path,
		hashAlgo:  hashAlgo,
		hashSize:  hashSize,
		packFiles: make(map[string]*packFile),
	}, nil
}

// Close closes the repository, releasing any resources associated with it.
//
// It is safe to call Close multiple times; subsequent calls will have no
// effect.
//
// Close invalidates any objects derived from the Repository as it;
// using them may cause segmentation faults or other undefined behavior.
func (repo *Repository) Close() error {
	var closeErr error
	repo.closeOnce.Do(func() {
		repo.packFilesMu.Lock()
		for key, pf := range repo.packFiles {
			err := pf.Close()
			if err != nil && closeErr == nil {
				closeErr = err
			}
			delete(repo.packFiles, key)
		}
		repo.packFilesMu.Unlock()
		if len(repo.packIdx) > 0 {
			for _, idx := range repo.packIdx {
				err := idx.Close()
				if err != nil && closeErr == nil {
					closeErr = err
				}
			}
		}
	})
	return closeErr
}

// repoPath joins the root with a relative path.
func (repo *Repository) repoPath(rel string) string {
	return filepath.Join(repo.rootPath, rel)
}

// ParseHash converts a hex string into a Hash, validating
// it matches the repository's hash size.
func (repo *Repository) ParseHash(s string) (Hash, error) {
	var id Hash
	if len(s)%2 != 0 {
		return id, fmt.Errorf("furgit: invalid hash length %d, it has to be even at the very least", len(s))
	}
	expectedLen := repo.hashSize * 2
	if len(s) != expectedLen {
		return id, fmt.Errorf("furgit: hash length mismatch: got %d chars, expected %d for hash size %d", len(s), expectedLen, repo.hashSize)
	}
	data, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("furgit: decode hash: %w", err)
	}
	copy(id.data[:], data)
	id.algo = repo.hashAlgo
	return id, nil
}

// computeRawHash computes a hash from raw data using the repository's hash algorithm.
func (repo *Repository) computeRawHash(data []byte) Hash {
	hashFunc := hashFuncs[repo.hashAlgo]
	return hashFunc(data)
}

// verifyRawObject verifies a raw object against its expected hash.
func (repo *Repository) verifyRawObject(buf []byte, want Hash) bool { //nolint:unused
	if want.algo != repo.hashAlgo {
		return false
	}
	return repo.computeRawHash(buf) == want
}

// verifyTypedObject verifies a typed object against its expected hash.
func (repo *Repository) verifyTypedObject(ty ObjectType, body []byte, want Hash) bool { //nolint:unused
	if want.algo != repo.hashAlgo {
		return false
	}
	header, err := headerForType(ty, body)
	if err != nil {
		return false
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return repo.computeRawHash(raw) == want
}
