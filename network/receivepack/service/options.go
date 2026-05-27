package service

import (
	"lindenii.org/go/furgit/common/iowrap"
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectstore "lindenii.org/go/furgit/object/store"
	refstore "lindenii.org/go/furgit/ref/store"
)

// Options configures one protocol-independent receive-pack service.
//
// Service borrows all configured dependencies.
//
// Refs and ExistingObjects are required and must be non-nil.
// ObjectIngress is required if Execute may need to ingest or quarantine a
// pack.
// CommitGraph, Progress, Hook, and HookIO are optional; when provided they are also
// borrowed for the duration of Execute.
type Options struct {
	Refs interface {
		refstore.Reader
		refstore.Transactioner
		refstore.Batcher
	}
	ExistingObjects objectstore.Reader
	ObjectIngress   objectstore.Quarantiner
	CommitGraph     *commitgraphread.Reader
	Progress        iowrap.WriteFlusher
	Hook            Hook
	HookIO          HookIO
}
