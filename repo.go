package furgit

import (
	"os"
	"path/filepath"
	"sync"
)

// Repository represents the root of a Git repository.
type Repository[T HashType] struct {
	rootPath string

	packIdxOnce sync.Once
	packIdx     []*packIndex[T]
	packIdxErr  error

	midxOnce sync.Once
	midx     *multiPackIndex[T]
	midxErr  error

	packFiles sync.Map // string, *packFile
	closeOnce sync.Once
}

// OpenRepository opens the repository at the provided path with the specified hash type.
// This will be replaced later with a function that auto-detects the hash size based
// on the git configuration.
func OpenRepository[T HashType](path string) (*Repository[T], error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrInvalidObject
	}
	return &Repository[T]{rootPath: path}, nil
}

// hashSize returns the hash size for this repository.
func (r *Repository[T]) hashSize() int {
	return hashLen[T]()
}

// ParseHash is a convenience method for parsing hashes in the context of this repository.
func (r *Repository[T]) ParseHash(s string) (Hash[T], error) {
	return ParseHash[T](s)
}

func (r *Repository[T]) Close() error {
	var closeErr error
	r.closeOnce.Do(func() {
		r.packFiles.Range(func(keya any, pfa any) bool {
			key := keya.(string)
			pf := pfa.(*packFile)
			err := pf.Close()
			if err != nil && closeErr == nil {
				closeErr = err
			}
			r.packFiles.Delete(key)
			return true
		})
		if len(r.packIdx) > 0 {
			for _, idx := range r.packIdx {
				err := idx.Close()
				if err != nil && closeErr == nil {
					closeErr = err
				}
			}
		}
		if r.midx != nil {
			err := r.midx.Close()
			if err != nil && closeErr == nil {
				closeErr = err
			}
		}
	})
	return closeErr
}

// Root returns the repository root path.
func (r *Repository[T]) Root() string {
	return r.rootPath
}

// repoPath joins the root with a relative path.
func (r *Repository[T]) repoPath(rel string) string {
	return filepath.Join(r.rootPath, rel)
}

func (r *Repository[T]) packFile(rel string) (*packFile, error) {
	if pf, ok := r.packFiles.Load(rel); ok {
		return pf.(*packFile), nil
	}
	pf, err := openPackFile(r.repoPath(rel), rel)
	if err != nil {
		return nil, err
	}
	actual, loaded := r.packFiles.LoadOrStore(rel, pf)
	if loaded {
		_ = pf.Close()
		return actual.(*packFile), nil
	}
	return pf, nil
}
