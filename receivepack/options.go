package receivepack

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

// Options configures one receive-pack invocation.
type Options struct {
	// GitProtocol is the raw Git protocol version string from the transport,
	// such as "version=1".
	GitProtocol string
	// Algorithm is the repository object ID algorithm used by the push session.
	Algorithm objectid.Algorithm
	// Refs is the reference store visible to the push.
	Refs refstore.ReadingStore
	// ExistingObjects is the object store visible to the push before any newly
	// uploaded quarantined objects are promoted.
	ExistingObjects objectstore.Store
	// ObjectsRoot is the permanent object storage root beneath which per-push
	// quarantine directories are derived.
	ObjectsRoot *os.Root
	// TODO: Hook and policy callbacks.
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
