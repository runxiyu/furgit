package receivepack

import (
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/ref/store"
)

// Options configures one receive-pack invocation.
//
// ReceivePack borrows all configured dependencies.
//
// Refs and ExistingObjects are required and must be non-nil.
// ObjectsRoot is required if the invocation may need to ingest or promote a
// pack.
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
	// quarantine promotion or ref updates. Hook is borrowed for the duration of
	// ReceivePack.
	Hook Hook
	// Agent is the receive-pack agent string advertised via capability.
	//
	// When empty, ReceivePack derives one from build info and falls back to
	// "furgit".
	Agent string
	// SessionID is the advertised receive-pack session-id capability value.
	//
	// When empty, ReceivePack generates one random value per invocation.
	SessionID string
	// PushCertNonce is the advertised push-cert nonce capability value.
	//
	// When empty, ReceivePack generates one random value per invocation.
	PushCertNonce string
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
