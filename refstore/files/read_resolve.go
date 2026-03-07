package files

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Resolve resolves one reference name from the files store visible namespace.
func (store *Store) Resolve(name string) (ref.Ref, error) { //nolint:ireturn
	if name == "" {
		return nil, refstore.ErrReferenceNotFound
	}

	resolved, err := store.readLooseRef(name)
	if err == nil {
		return resolved, nil
	}

	if !errors.Is(err, refstore.ErrReferenceNotFound) {
		refPath := store.loosePath(name)

		info, statErr := store.rootFor(refPath.root).Stat(refPath.path)
		if statErr != nil || !info.IsDir() {
			return nil, err
		}
	}

	packed, packedErr := store.readPackedRefs()
	if packedErr != nil {
		return nil, packedErr
	}

	detached, ok := packed.byName[name]
	if !ok {
		return nil, refstore.ErrReferenceNotFound
	}

	return detached, nil
}

// ResolveFully resolves symbolic references through the visible files store
// namespace until one detached reference is reached.
func (store *Store) ResolveFully(name string) (ref.Detached, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return ref.Detached{}, fmt.Errorf("refstore/files: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		resolved, err := store.Resolve(cur)
		if err != nil {
			return ref.Detached{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Detached:
			return resolved, nil
		case ref.Symbolic:
			target := strings.TrimSpace(resolved.Target)
			if target == "" {
				return ref.Detached{}, fmt.Errorf("refstore/files: symbolic reference %q has empty target", resolved.Name())
			}

			cur = target
		default:
			return ref.Detached{}, fmt.Errorf("refstore/files: unsupported reference type %T", resolved)
		}
	}
}
