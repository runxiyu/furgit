package service

import (
	"io"
	"io/fs"
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

type PromotedObjectPermissions struct {
	DirMode  fs.FileMode
	FileMode fs.FileMode
}

// Options configures one protocol-independent receive-pack service.
//
// Service borrows all configured dependencies.
//
// Refs and ExistingObjects are required and must be non-nil.
// ObjectsRoot is required if Execute may need to ingest or promote a pack.
// Progress, ProgressFlush, Hook, and HookIO are optional; when provided they
// are also borrowed for the duration of Execute.
type Options struct {
	Algorithm                 objectid.Algorithm
	Refs                      refstore.ReadWriteStore
	ExistingObjects           objectstorer.Store
	ObjectsRoot               *os.Root
	Progress                  io.Writer
	ProgressFlush             func() error
	PromotedObjectPermissions *PromotedObjectPermissions
	Hook                      Hook
	HookIO                    HookIO
}
