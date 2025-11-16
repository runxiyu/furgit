package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Repository represents the root of a Git repository.
type Repository struct {
	rootPath string
	HashSize int

	packIdxOnce sync.Once
	packIdx     []*packIndex
	packIdxErr  error

	midxOnce sync.Once
	midx     *multiPackIndex
	midxErr  error

	packFiles sync.Map // string, *packFile
	closeOnce sync.Once
}

// OpenRepository opens the repository at the provided path. The path is expected to be
// the actual repository directory, i.e., the repository itself for bare repositories,
// or the .git subdirectory for non-bare repositories.
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

	cfg, err := ParseConfig(f)
	if err != nil {
		return nil, fmt.Errorf("furgit: failed to parse config: %w", err)
	}

	algo := cfg.Get("extensions", "", "objectformat")
	if algo == "" {
		algo = "sha1"
	}

	var hashSize int
	switch algo {
	case "sha1":
		hashSize = sha1.Size
	case "sha256":
		hashSize = sha256.Size
	default:
		return nil, fmt.Errorf("furgit: unsupported hash algorithm %q", algo)
	}

	if _, ok := hashFuncs[hashSize]; !ok {
		return nil, fmt.Errorf("furgit: hash algorithm %q is not supported by the hash functions provided by this build", algo)
	}

	return &Repository{rootPath: path, HashSize: hashSize}, nil
}

func (repo *Repository) Close() error {
	var closeErr error
	repo.closeOnce.Do(func() {
		repo.packFiles.Range(func(keya any, pfa any) bool {
			key := keya.(string)
			pf := pfa.(*packFile)
			err := pf.Close()
			if err != nil && closeErr == nil {
				closeErr = err
			}
			repo.packFiles.Delete(key)
			return true
		})
		if len(repo.packIdx) > 0 {
			for _, idx := range repo.packIdx {
				err := idx.Close()
				if err != nil && closeErr == nil {
					closeErr = err
				}
			}
		}
		if repo.midx != nil {
			err := repo.midx.Close()
			if err != nil && closeErr == nil {
				closeErr = err
			}
		}
	})
	return closeErr
}

// Root returns the repository root path.
func (repo *Repository) Root() string {
	return repo.rootPath
}

// repoPath joins the root with a relative path.
func (repo *Repository) repoPath(rel string) string {
	return filepath.Join(repo.rootPath, rel)
}

func (repo *Repository) packFile(rel string) (*packFile, error) {
	if pf, ok := repo.packFiles.Load(rel); ok {
		return pf.(*packFile), nil
	}
	pf, err := openPackFile(repo.repoPath(rel), rel)
	if err != nil {
		return nil, err
	}
	actual, loaded := repo.packFiles.LoadOrStore(rel, pf)
	if loaded {
		_ = pf.Close()
		return actual.(*packFile), nil
	}
	return pf, nil
}

// ParseHash converts a hex string into a Hash, validating it matches the repository's hash size.
func (repo *Repository) ParseHash(s string) (Hash, error) {
	var id Hash
	if len(s)%2 != 0 {
		return id, fmt.Errorf("furgit: invalid hash length %d, it has to be even at the very least", len(s))
	}
	expectedLen := repo.HashSize * 2
	if len(s) != expectedLen {
		return id, fmt.Errorf("furgit: hash length mismatch: got %d chars, expected %d for hash size %d", len(s), expectedLen, repo.HashSize)
	}
	data, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("furgit: decode hash: %w", err)
	}
	copy(id.data[:], data)
	id.size = len(s) / 2
	return id, nil
}

// computeRawHash computes a hash from raw data using the repository's hash algorithm.
func (repo *Repository) computeRawHash(data []byte) Hash {
	hashFunc := hashFuncs[repo.HashSize]
	return hashFunc(data)
}

// verifyRawObject verifies a raw object against its expected hash.
func (repo *Repository) verifyRawObject(buf []byte, want Hash) bool {
	if want.size != repo.HashSize {
		return false
	}
	return repo.computeRawHash(buf) == want
}

// verifyTypedObject verifies a typed object against its expected hash.
func (repo *Repository) verifyTypedObject(ty ObjType, body []byte, want Hash) bool {
	if want.size != repo.HashSize {
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
