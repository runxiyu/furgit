package service

import (
	"codeberg.org/lindenii/furgit/common/iowrap"
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	refstore "codeberg.org/lindenii/furgit/ref/store"
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
	Algorithm objectid.Algorithm
	Refs      interface {
		refstore.ReadingStore
		refstore.TransactionalStore
		refstore.BatchStore
	}
	ExistingObjects objectstore.Reader
	ObjectIngress   objectstore.Quarantiner
	CommitGraph     *commitgraphread.Reader
	Progress        iowrap.WriteFlusher
	Hook            Hook
	HookIO          HookIO
}
