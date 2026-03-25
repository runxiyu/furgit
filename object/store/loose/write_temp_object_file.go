package loose

import (
	"crypto/rand"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// createTempObjectFile creates a unique temporary object file within dir.
// The returned path is relative to the objects root.
func (store *Store) createTempObjectFile(dir string) (string, *os.File, error) {
	for range 16 {
		relPath := filepath.Join(dir, tempObjectFilePrefix+rand.Text())

		file, err := store.root.OpenFile(relPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return relPath, file, nil
		}

		if errors.Is(err, fs.ErrExist) {
			continue
		}

		return "", nil, err
	}

	return "", nil, errors.New("objectstore/loose: failed to create temporary object file")
}
