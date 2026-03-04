package loose

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Resolve resolves a loose reference name to symbolic or detached form.
func (store *Store) Resolve(name string) (ref.Ref, error) {
	if name == "" {
		return nil, refstore.ErrReferenceNotFound
	}

	resolved, err := store.resolveOne(name)
	if err != nil {
		return nil, err
	}

	return resolved, nil
}

// ResolveFully resolves symbolic references within the loose backend only.
func (store *Store) ResolveFully(name string) (ref.Detached, error) {
	if name == "" {
		return ref.Detached{}, refstore.ErrReferenceNotFound
	}

	cur := name

	seen := make(map[string]struct{})
	for {
		if _, ok := seen[cur]; ok {
			return ref.Detached{}, fmt.Errorf("refstore/loose: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		resolved, err := store.resolveOne(cur)
		if err != nil {
			return ref.Detached{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Detached:
			return resolved, nil
		case ref.Symbolic:
			target := strings.TrimSpace(resolved.Target)
			if target == "" {
				return ref.Detached{}, fmt.Errorf("refstore/loose: symbolic reference %q has empty target", resolved.Name())
			}

			cur = target
		default:
			return ref.Detached{}, fmt.Errorf("refstore/loose: unsupported reference type %T", resolved)
		}
	}
}

// resolveOne resolves one loose ref file without symbolic recursion.
func (store *Store) resolveOne(name string) (ref.Ref, error) {
	data, err := store.root.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, refstore.ErrReferenceNotFound
		}

		return nil, err
	}

	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: ") {
		target := strings.TrimSpace(line[len("ref: "):])
		if target == "" {
			return nil, fmt.Errorf("refstore/loose: symbolic reference %q has empty target", name)
		}

		return ref.Symbolic{
			RefName: name,
			Target:  target,
		}, nil
	}

	id, err := objectid.ParseHex(store.algo, line)
	if err != nil {
		return nil, fmt.Errorf("refstore/loose: invalid detached reference %q: %w", name, err)
	}

	return ref.Detached{
		RefName: name,
		ID:      id,
	}, nil
}
