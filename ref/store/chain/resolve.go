package chain

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/ref"
	refstore "lindenii.org/go/furgit/ref/store"
)

// Resolve resolves a reference from the first backend that has it.
//
//nolint:ireturn
func (chain *Chain) Resolve(name string) (ref.Ref, error) {
	for i, backend := range chain.backends {
		resolved, err := backend.Resolve(name)
		if err == nil {
			return resolved, nil
		}

		if errors.Is(err, refstore.ErrReferenceNotFound) {
			continue
		}

		return nil, fmt.Errorf("refstore: backend %d resolve: %w", i, err)
	}

	return nil, refstore.ErrReferenceNotFound
}

// ResolveToDetached resolves symbolic references through Resolve until detached.
//
// It intentionally does not call backend ResolveToDetached. This allows symbolic
// references to cross backends in the chain.
func (chain *Chain) ResolveToDetached(name string) (ref.Detached, error) {
	cur := name

	seen := map[string]struct{}{}
	for {
		if _, ok := seen[cur]; ok {
			return ref.Detached{}, fmt.Errorf("refstore: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		resolved, err := chain.Resolve(cur)
		if err != nil {
			return ref.Detached{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Detached:
			return resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Detached{}, fmt.Errorf("refstore: symbolic reference %q has empty target", resolved.Name())
			}

			cur = resolved.Target
		default:
			return ref.Detached{}, fmt.Errorf("refstore: unsupported reference type %T", resolved)
		}
	}
}
