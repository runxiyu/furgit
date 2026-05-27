// Package loose provides a loose object backend (objects/XX/YYYYY..).
package loose

import (
	"os"

	objectid "lindenii.org/go/furgit/object/id"
)

// Store reads loose Git objects from an objects directory root.
//
// Loose objects are zlib streams whose trailer uses Adler-32. Which reads
// consume enough of the stream to reach and verify that trailer is documented
// on the individual methods.
//
// Labels: Close-Caller.
type Store struct {
	// root is the objects directory capability used for all object file access.
	// Object files are opened by relative paths like "<first2>/<rest>".
	// Store borrows this root.
	root *os.Root
	// algo is the expected object ID algorithm for lookups.
	algo objectid.Algorithm
}

// New creates a loose-object store rooted at an objects directory for algo.
//
// Labels: Deps-Borrowed, Life-Parent.
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
// Labels: MT-Unsafe.
func (store *Store) Close() error { return nil }
