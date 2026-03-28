package hooks

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/commitquery"
	receivepack "codeberg.org/lindenii/furgit/network/receivepack"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectmix "codeberg.org/lindenii/furgit/object/store/mix"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// RejectForcePush rejects updates whose new value is not a fast-forward of the
// currently resolved reference.
func RejectForcePush() receivepack.Hook {
	return func(
		ctx context.Context,
		req receivepack.HookRequest,
	) ([]receivepack.UpdateDecision, error) {
		_ = ctx

		objects := objectmix.New(req.QuarantinedObjects, req.ExistingObjects)

		decisions := make([]receivepack.UpdateDecision, len(req.Updates))
		for i := range decisions {
			decisions[i].Accept = true
		}

		for i, update := range req.Updates {
			if update.OldID == objectid.Zero(update.OldID.Algorithm()) || update.NewID == objectid.Zero(update.NewID.Algorithm()) {
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

			ok, err := commitquery.New(objects, req.CommitGraph).IsAncestor(current.ID, update.NewID)
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
