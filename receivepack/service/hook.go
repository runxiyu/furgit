package service

import (
	"context"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/ref/store"
)

type HookIO struct {
	Progress io.Writer
	Error    io.Writer
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

// HookRequest is the borrowed view passed to one Hook invocation.
//
// Refs, ExistingObjects, and QuarantinedObjects are borrowed and are only
// valid for the duration of the hook call.
type HookRequest struct {
	Refs               refstore.ReadingStore
	ExistingObjects    objectstore.Store
	QuarantinedObjects objectstore.Store
	Updates            []RefUpdate
	PushOptions        []string
	IO                 HookIO
}

// Hook is an optional per-request validation hook.
//
// Hook borrows the data and stores in HookRequest only for the duration of the
// call.
type Hook func(context.Context, HookRequest) ([]UpdateDecision, error)
