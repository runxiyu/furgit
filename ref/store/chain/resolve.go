package chain

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
)

// Resolve resolves one reference name
// from the first backend that has it.
func (chain *Chain) Resolve(name string) (ref.Ref, error) {
	for i, backend := range chain.backends {
		resolved, err := backend.Resolve(name)
		if err == nil {
			return resolved, nil
		}

		if errors.Is(err, store.ErrReferenceNotFound) {
			continue
		}

		return nil, fmt.Errorf("ref/store/chain: backend %d resolve: %w", i, err)
	}

	return nil, store.ErrReferenceNotFound
}

// ResolveToDirect resolves symbolic references
// until one direct reference is reached.
//
// It follows each symbolic hop through the whole chain
// rather than through a single backend,
// so a symbolic reference in one backend
// may target a reference in another.
func (chain *Chain) ResolveToDirect(name string) (ref.Direct, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return ref.Direct{}, fmt.Errorf("%w: at %q", store.ErrSymbolicCycle, cur)
		}

		seen[cur] = struct{}{}

		resolved, err := chain.Resolve(cur)
		if err != nil {
			return ref.Direct{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Direct:
			return resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Direct{}, fmt.Errorf(
					"%w: symbolic reference %q has empty target",
					store.ErrInvalidValue, resolved.Name(),
				)
			}

			cur = resolved.Target
		default:
			panic(fmt.Sprintf("ref/store/chain: unsupported reference type %T", resolved))
		}
	}
}
