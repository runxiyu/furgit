package receivepack

import (
	"context"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/refstore"
)

// RefUpdate is one requested reference update presented to a receive-pack hook.
type RefUpdate struct {
	Name  string
	OldID objectid.ObjectID
	NewID objectid.ObjectID
}

// UpdateDecision is one hook decision for a requested reference update.
type UpdateDecision struct {
	Accept  bool
	Message string
}

// HookRequest is the input presented to a receive-pack hook before quarantine
// promotion and ref updates.
type HookRequest struct {
	Refs               refstore.ReadingStore
	ExistingObjects    objectstore.Store
	QuarantinedObjects objectstore.Store
	Updates            []RefUpdate
	PushOptions        []string
}

// Hook decides whether each requested update should proceed.
//
// The hook runs after pack ingestion into quarantine and before quarantine
// promotion or ref updates. The returned decisions must have the same length as
// HookRequest.Updates.
type Hook func(context.Context, HookRequest) ([]UpdateDecision, error)
