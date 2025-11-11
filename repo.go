package furgit

import (
	"os"
	"path/filepath"
	"sync"
)

// Repository represents the root of a Git repository.
type Repository struct {
	rootPath string

	packIdxOnce sync.Once
	packIdx     []*packIndex
	packIdxErr  error

	packFiles sync.Map // string, *packFile
	closeOnce sync.Once
}

// OpenRepository opens the repository at the provided path.
func OpenRepository(path string) (*Repository, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrInvalidObject
	}
	return &Repository{rootPath: path}, nil
}

func (r *Repository) Close() error {
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
	})
	return closeErr
}

// Root returns the repository root path.
func (r *Repository) Root() string {
	return r.rootPath
}

// repoPath joins the root with a relative path.
func (r *Repository) repoPath(rel string) string {
	return filepath.Join(r.rootPath, rel)
}

func (r *Repository) packFile(rel string) (*packFile, error) {
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
