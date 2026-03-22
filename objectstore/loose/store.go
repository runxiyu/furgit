// Package loose provides a loose object backend (objects/XX/YYYYY..).
package loose

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
)

// Store reads loose Git objects from an objects directory root.
//
// Loose objects are zlib streams whose trailer uses Adler-32. Which reads
// consume enough of the stream to reach and verify that trailer is documented
// on the individual methods.
//
// Store owns root and closes it in Close.
type Store struct {
	// root is the objects directory capability used for all object file access.
	// Object files are opened by relative paths like "<first2>/<rest>".
	// Store owns this root.
	root *os.Root
	// algo is the expected object ID algorithm for lookups.
	algo objectid.Algorithm
}

// New creates a loose-object store rooted at an objects directory for algo.
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
//
// Repeated calls to Close are undefined behavior.
func (store *Store) Close() error {
	return store.root.Close()
}
