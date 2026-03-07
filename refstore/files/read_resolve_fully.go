package files

import (
	"fmt"
	"strings"

	"codeberg.org/lindenii/furgit/ref"
)

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
