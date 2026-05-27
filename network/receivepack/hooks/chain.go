package hooks

import (
	"context"
	"fmt"

	receivepack "lindenii.org/go/furgit/network/receivepack"
)

// Chain combines hooks by running them in order and intersecting their
// decisions. The first rejecting message for each update is preserved.
func Chain(hooks ...receivepack.Hook) receivepack.Hook {
	return func(
		ctx context.Context,
		req receivepack.HookRequest,
	) ([]receivepack.UpdateDecision, error) {
		decisions := make([]receivepack.UpdateDecision, len(req.Updates))
		for i := range decisions {
			decisions[i].Accept = true
		}

		for _, hook := range hooks {
			if hook == nil {
				continue
			}

			hookDecisions, err := hook(ctx, req)
			if err != nil {
				return nil, err
			}

			if len(hookDecisions) != len(req.Updates) {
				return nil, fmt.Errorf("hook returned %d decisions for %d updates", len(hookDecisions), len(req.Updates))
			}

			for i, decision := range hookDecisions {
				if decision.Accept {
					continue
				}

				if decisions[i].Accept {
					decisions[i].Message = decision.Message
				}

				decisions[i].Accept = false
			}
		}

		return decisions, nil
	}
}
