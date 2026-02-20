// Package loose provides loose-object reads from a repository root.
package loose

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
)

// Store reads loose Git objects from a repository root.
//
// Store does not own root. Callers are responsible for closing root.
type Store struct {
	// root is the repository root capability used for all object file access.
	// Store does not own this root.
	root *os.Root
	// algo is the expected object ID algorithm for lookups.
	algo objectid.Algorithm
}

// New creates a loose-object store rooted at root for algo.
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
	_ = store
	return nil
}
