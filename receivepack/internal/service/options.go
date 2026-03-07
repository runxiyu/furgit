package service

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

// Options configures one protocol-independent receive-pack service.
type Options struct {
	Algorithm       objectid.Algorithm
	Refs            refstore.ReadingStore
	ExistingObjects objectstore.Store
	ObjectsRoot     *os.Root
	// TODO: Hook and such callbacks.
}
