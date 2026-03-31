package service

import (
	"context"

	"codeberg.org/lindenii/furgit/common/iowrap"
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

type HookIO struct {
	Progress iowrap.WriteFlusher
	Error    iowrap.WriteFlusher
}

type RefUpdate struct {
	Name  string
	OldID objectid.ObjectID
	NewID objectid.ObjectID
}

type UpdateDecision struct {
	Accept  bool
	Message string
}

// HookRequest is the view passed to one Hook invocation.
//
// Labels: Life-Call.
type HookRequest struct {
	Refs            refstore.Reader
	ExistingObjects objectstore.Reader
	// QuarantinedObjects exposes quarantined objects for this push.
	//
	// When the push did not create a quarantine, QuarantinedObjects is nil.
	QuarantinedObjects objectstore.Reader
	CommitGraph        *commitgraphread.Reader
	Updates            []RefUpdate
	PushOptions        []string
	IO                 HookIO
}

// Hook is an optional per-request validation hook.
//
// The returned decisions must have the same length as HookRequest.Updates.
type Hook func(context.Context, HookRequest) ([]UpdateDecision, error)
