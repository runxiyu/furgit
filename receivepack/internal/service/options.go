package service

import (
	"io/fs"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

type PromotedObjectPermissions struct {
	DirMode  fs.FileMode
	FileMode fs.FileMode
}

// Options configures one protocol-independent receive-pack service.
type Options struct {
	Algorithm                 objectid.Algorithm
	Refs                      refstore.ReadWriteStore
	ExistingObjects           objectstore.Store
	ObjectsRoot               *os.Root
	PromotedObjectPermissions *PromotedObjectPermissions
	Hook                      Hook
}
