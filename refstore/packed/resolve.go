package packed

import (
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Resolve resolves a packed reference name to a detached ref.
//
//nolint:ireturn
func (store *Store) Resolve(name string) (ref.Ref, error) {
	detached, ok := store.byName[name]
	if !ok {
		return nil, refstore.ErrReferenceNotFound
	}

	return detached, nil
}

// ResolveFully resolves a packed reference name to a detached ref.
//
// Packed refs are detached-only, so ResolveFully is equivalent to Resolve.
func (store *Store) ResolveFully(name string) (ref.Detached, error) {
	detached, ok := store.byName[name]
	if !ok {
		return ref.Detached{}, refstore.ErrReferenceNotFound
	}

	return detached, nil
}
