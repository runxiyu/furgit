package receivepack

import (
	"io/fs"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

// PromotedObjectPermissions configures the destination permissions applied to
// objects and directories promoted out of quarantine.
type PromotedObjectPermissions struct {
	DirMode  fs.FileMode
	FileMode fs.FileMode
}

// Options configures one receive-pack invocation.
type Options struct {
	// GitProtocol is the raw Git protocol version string from the transport,
	// such as "version=1".
	GitProtocol string
	// Algorithm is the repository object ID algorithm used by the push session.
	Algorithm objectid.Algorithm
	// Refs is the reference store visible to the push.
	Refs refstore.ReadWriteStore
	// ExistingObjects is the object store visible to the push before any newly
	// uploaded quarantined objects are promoted.
	ExistingObjects objectstore.Store
	// ObjectsRoot is the permanent object storage root beneath which per-push
	// quarantine directories are derived.
	ObjectsRoot *os.Root
	// PromotedObjectPermissions, when non-nil, is applied to objects and
	// directories moved from quarantine into the permanent object store.
	PromotedObjectPermissions *PromotedObjectPermissions
	// Hook, when non-nil, runs after pack ingestion into quarantine and before
	// quarantine promotion or ref updates.
	Hook Hook
}

func validateOptions(opts Options) error {
	if opts.Algorithm == 0 {
		return ErrMissingAlgorithm
	}

	if opts.Refs == nil {
		return ErrMissingRefs
	}

	if opts.ExistingObjects == nil {
		return ErrMissingObjects
	}

	return nil
}
