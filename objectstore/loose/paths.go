package loose

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// objectPath returns the loose object path for id relative to the objects root.
func (store *Store) objectPath(id objectid.ObjectID) (string, error) {
	if id.Algorithm() != store.algo {
		return "", fmt.Errorf("objectstore/loose: object id algorithm mismatch: got %s want %s", id.Algorithm(), store.algo)
	}
	hex := id.String()
	return filepath.Join(hex[:2], hex[2:]), nil
}

// openObject opens the loose object file for id.
// Missing files cause objectstore.ErrObjectNotFound.
func (store *Store) openObject(id objectid.ObjectID) (*os.File, error) {
	relPath, err := store.objectPath(id)
	if err != nil {
		return nil, err
	}
	file, err := store.root.Open(relPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, objectstore.ErrObjectNotFound
		}
		return nil, err
	}
	return file, nil
}
