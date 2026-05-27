package service

import (
	"context"

	"lindenii.org/go/furgit/common/iowrap"
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	refstore "lindenii.org/go/furgit/ref/store"
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
