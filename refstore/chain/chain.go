// Package chain provides an ordered reference store chain implementation.
package chain

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Chain queries multiple reference stores in order.
type Chain struct {
	backends []refstore.Store
}

// New creates an ordered reference store chain.
func New(backends ...refstore.Store) *Chain {
	return &Chain{
		backends: append([]refstore.Store(nil), backends...),
	}
}

// Resolve resolves a reference from the first backend that has it.
func (chain *Chain) Resolve(name string) (ref.Ref, error) {
	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
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

// ResolveFully resolves symbolic references through Resolve until detached.
//
// It intentionally does not call backend ResolveFully. This allows symbolic
// references to cross backends in the chain.
func (chain *Chain) ResolveFully(name string) (ref.Detached, error) {
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
		case *ref.Detached:
			if resolved == nil {
				return ref.Detached{}, fmt.Errorf("refstore: backend returned nil detached reference")
			}
			return *resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Detached{}, fmt.Errorf("refstore: symbolic reference %q has empty target", resolved.Name())
			}
			cur = resolved.Target
		case *ref.Symbolic:
			if resolved == nil {
				return ref.Detached{}, fmt.Errorf("refstore: backend returned nil symbolic reference")
			}
			if resolved.Target == "" {
				return ref.Detached{}, fmt.Errorf("refstore: symbolic reference %q has empty target", resolved.Name())
			}
			cur = resolved.Target
		default:
			return ref.Detached{}, fmt.Errorf("refstore: unsupported reference type %T", resolved)
		}
	}
}

// List lists references from every backend and deduplicates by ref name.
//
// First-seen wins, so earlier backends have precedence.
func (chain *Chain) List(pattern string) ([]ref.Ref, error) {
	var refs []ref.Ref
	seen := map[string]struct{}{}

	for i, backend := range chain.backends {
		if backend == nil {
			continue
		}
		listed, err := backend.List(pattern)
		if err != nil {
			return nil, fmt.Errorf("refstore: backend %d list: %w", i, err)
		}
		for _, entry := range listed {
			if entry == nil {
				continue
			}
			name := entry.Name()
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			refs = append(refs, entry)
		}
	}

	return refs, nil
}

// Close closes all backends and joins close errors.
func (chain *Chain) Close() error {
	var errs []error
	for _, backend := range chain.backends {
		if backend == nil {
			continue
		}
		if err := backend.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
