package hooks

import (
	"context"
	"errors"
	"fmt"

	"lindenii.org/go/furgit/commitquery"
	receivepack "lindenii.org/go/furgit/network/receivepack"
	"lindenii.org/go/furgit/object/fetch"
	objectmix "lindenii.org/go/furgit/object/store/mix"
	refstore "lindenii.org/go/furgit/ref/store"
)

// RejectForcePush rejects updates whose new value is not a fast-forward of the
// currently resolved reference.
func RejectForcePush() receivepack.Hook {
	return func(
		ctx context.Context,
		req receivepack.HookRequest,
	) ([]receivepack.UpdateDecision, error) {
		_ = ctx

		objects := req.ExistingObjects
		if req.QuarantinedObjects != nil {
			objects = objectmix.New(req.QuarantinedObjects, req.ExistingObjects)
		}

		queries := commitquery.New(fetch.New(objects), req.CommitGraph)

		decisions := make([]receivepack.UpdateDecision, len(req.Updates))
		for i := range decisions {
			decisions[i].Accept = true
		}

		for i, update := range req.Updates {
			if update.OldID == update.OldID.Algorithm().Zero() || update.NewID == update.NewID.Algorithm().Zero() {
				continue
			}

			current, err := req.Refs.ResolveToDetached(update.Name)
			switch {
			case err == nil:
			case errors.Is(err, refstore.ErrReferenceNotFound):
				continue
			default:
				return nil, fmt.Errorf("resolve %s: %w", update.Name, err)
			}

			if current.ID == update.NewID {
				continue
			}

			ok, err := queries.IsAncestor(current.ID, update.NewID)
			if err != nil {
				return nil, fmt.Errorf("check fast-forward %s: %w", update.Name, err)
			}

			if !ok {
				decisions[i] = receivepack.UpdateDecision{
					Accept:  false,
					Message: "non-fast-forward",
				}
			}
		}

		return decisions, nil
	}
}
