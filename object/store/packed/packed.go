package packed

import (
	"os"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

// Packed reads Git objects from pack/index files under an objects/pack root,
// and ingests incoming pack streams into it.
//
// Labels: Close-Caller.
type Packed struct {
	// root is the objects/pack directory capability
	// used for all pack and index file access.
	// Packed borrows this root.
	root *os.Root
	// objectFormat is the expected object format for lookups.
	objectFormat id.ObjectFormat
}

var (
	_ store.ObjectReader = (*Packed)(nil)
	_ store.PackWriter   = (*Packed)(nil)
)

// New creates a packed-object store rooted at an objects/pack directory.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, objectFormat id.ObjectFormat) (*Packed, error)

// Close releases mapped pack/index resources associated with the store.
//
// Labels: MT-Unsafe.
func (packed *Packed) Close() error
