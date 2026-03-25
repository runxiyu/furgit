package receivepack

import (
	"context"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/receivepack/service"
	"codeberg.org/lindenii/furgit/refstore"
)

type HookIO struct {
	Progress io.Writer
	Error    io.Writer
}

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

// Hook decides whether each requested update should proceed.
//
// The hook runs after pack ingestion into quarantine and before quarantine
// promotion or ref updates. The returned decisions must have the same length as
// HookRequest.Updates. Hook borrows the data and stores in HookRequest only for
// the duration of the call.
type Hook func(context.Context, HookRequest) ([]UpdateDecision, error)

func translateHook(hook Hook) service.Hook {
	if hook == nil {
		return nil
	}

	return func(ctx context.Context, req service.HookRequest) ([]service.UpdateDecision, error) {
		translatedUpdates := make([]RefUpdate, 0, len(req.Updates))
		for _, update := range req.Updates {
			translatedUpdates = append(translatedUpdates, RefUpdate{
				Name:  update.Name,
				OldID: update.OldID,
				NewID: update.NewID,
			})
		}

		decisions, err := hook(ctx, HookRequest{
			Refs:               req.Refs,
			ExistingObjects:    req.ExistingObjects,
			QuarantinedObjects: req.QuarantinedObjects,
			Updates:            translatedUpdates,
			PushOptions:        append([]string(nil), req.PushOptions...),
			IO: HookIO{
				Progress: req.IO.Progress,
				Error:    req.IO.Error,
			},
		})
		if err != nil {
			return nil, err
		}

		out := make([]service.UpdateDecision, 0, len(decisions))
		for _, decision := range decisions {
			out = append(out, service.UpdateDecision{
				Accept:  decision.Accept,
				Message: decision.Message,
			})
		}

		return out, nil
	}
}
