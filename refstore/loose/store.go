// Package loose provides a loose ref backend (refs/... as a directory tree).
package loose

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads loose references from a repository root.
//
// Store owns root and closes it in Close.
type Store struct {
	// root is the repository root capability.
	root *os.Root
	// algo is the object ID algorithm used by this repository.
	algo objectid.Algorithm
}

var _ refstore.Store = (*Store)(nil)

// New creates a loose ref store rooted at a repository root.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &Store{
		root: root,
		algo: algo,
	}, nil
}

// Close releases resources associated with the backend.
func (store *Store) Close() error {
	return store.root.Close()
}
