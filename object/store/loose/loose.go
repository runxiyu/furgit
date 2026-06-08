package loose

import (
	"fmt"
	"os"
	"path/filepath"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

// Loose reads loose Git objects from an objects directory root.
//
// Loose objects are zlib streams whose trailer uses Adler-32.
// Which reads consume enough of the stream
// to reach and verify that trailer
// is documented on the individual methods.
//
// Labels: Close-Caller.
type Loose struct {
	// root is the objects directory capability used for all object file access.
	// Object files are opened by relative paths like "<first2>/<rest>".
	// Loose borrows this root.
	root *os.Root
	// objectFormat is the expected object format for lookups.
	objectFormat id.ObjectFormat
}

var (
	_ store.ObjectReader = (*Loose)(nil)
	_ store.ObjectWriter = (*Loose)(nil)
)

// New creates a loose-object store rooted at an objects directory for objectFormat.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, objectFormat id.ObjectFormat) (*Loose, error) {
	if objectFormat.Size() == 0 {
		return nil, id.ErrInvalidObjectFormat
	}

	return &Loose{
		root:         root,
		objectFormat: objectFormat,
	}, nil
}

// Close releases resources associated with the backend.
//
// Labels: MT-Unsafe.
func (loose *Loose) Close() error { return nil }

// objectPath returns the loose object path for objectID relative to the objects root.
func (loose *Loose) objectPath(objectID id.ObjectID) (string, error) {
	if objectID.ObjectFormat() != loose.objectFormat {
		return "", fmt.Errorf("%w: got %s want %s", id.ErrInvalidObjectFormat, objectID.ObjectFormat(), loose.objectFormat)
	}

	hex := objectID.String()

	return filepath.Join(hex[:2], hex[2:]), nil
}
