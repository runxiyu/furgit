package memory

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"lindenii.org/go/furgit/ref"
	refstore "lindenii.org/go/furgit/ref/store"
)

// Resolve resolves one reference name
// from the in-memory namespace.
func (store *Store) Resolve(name string) (ref.Ref, error) { //nolint:ireturn
	store.mu.RLock()
	defer store.mu.RUnlock()

	return publicRef(name, store.refs[name])
}

// ResolveToDetached resolves symbolic references
// through the in-memory namespace
// until one detached reference is reached.
func (store *Store) ResolveToDetached(name string) (ref.Detached, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return store.resolveToDetachedLocked(name)
}

// List lists references from the in-memory namespace.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	matchAll := pattern == ""
	if !matchAll {
		_, err := path.Match(pattern, "HEAD")
		if err != nil {
			return nil, err
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	names := make([]string, 0, len(store.refs))
	for name := range store.refs {
		if !matchAll {
			matched, err := path.Match(pattern, name)
			if err != nil {
				return nil, err
			}

			if !matched {
				continue
			}
		}

		names = append(names, name)
	}

	slices.Sort(names)

	refs := make([]ref.Ref, 0, len(names))
	for _, name := range names {
		resolved, err := publicRef(name, store.refs[name])
		if err != nil {
			return nil, err
		}

		refs = append(refs, resolved)
	}

	return refs, nil
}

func (store *Store) resolveToDetachedLocked(name string) (ref.Detached, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return ref.Detached{}, fmt.Errorf("refstore/memory: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		resolved, err := publicRef(cur, store.refs[cur])
		if err != nil {
			return ref.Detached{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Detached:
			return resolved, nil
		case ref.Symbolic:
			target := strings.TrimSpace(resolved.Target)
			if target == "" {
				return ref.Detached{}, fmt.Errorf("refstore/memory: symbolic reference %q has empty target", resolved.Name())
			}

			cur = target
		default:
			return ref.Detached{}, fmt.Errorf("refstore/memory: unsupported reference type %T", resolved)
		}
	}
}

func publicRef(name string, stored storedRef) (ref.Ref, error) { //nolint:ireturn
	switch stored.kind {
	case storedDetached:
		detached := ref.Detached{RefName: name, ID: stored.id}
		if stored.peeled != nil {
			peeled := *stored.peeled
			detached.Peeled = &peeled
		}

		return detached, nil
	case storedSymbolic:
		return ref.Symbolic{RefName: name, Target: stored.target}, nil
	case storedMissing:
		return nil, refstore.ErrReferenceNotFound
	default:
		return nil, fmt.Errorf("refstore/memory: unsupported stored reference kind %d", stored.kind)
	}
}
