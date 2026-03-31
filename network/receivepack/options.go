package receivepack

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Options configures one receive-pack invocation.
//
// ReceivePack borrows all configured dependencies.
//
// Refs and ExistingObjects are required and must be non-nil.
// ObjectIngress is required if the invocation may need to ingest or quarantine a
// pack.
type Options struct {
	// GitProtocol is the raw Git protocol version string from the transport,
	// such as "version=1".
	GitProtocol string
	// Algorithm is the repository object ID algorithm used by the push session.
	Algorithm objectid.Algorithm
	// Refs is the reference store visible to the push.
	Refs interface {
		refstore.Reader
		refstore.Transactioner
		refstore.Batcher
	}
	// ExistingObjects is the object store visible to the push before any newly
	// uploaded quarantined objects are promoted.
	ExistingObjects objectstore.Reader
	// ObjectIngress creates coordinated quarantines for quarantined object and
	// pack ingestion during the push.
	ObjectIngress objectstore.Quarantiner
	// CommitGraph is an optional commit-graph snapshot corresponding to
	// ExistingObjects.
	CommitGraph *commitgraphread.Reader
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
