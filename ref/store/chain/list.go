package chain

import (
	"fmt"

	"lindenii.org/go/furgit/ref"
)

// List lists references from every backend and deduplicates by ref name.
//
// First-seen wins, so earlier backends have precedence.
func (chain *Chain) List(pattern string) ([]ref.Ref, error) {
	var refs []ref.Ref

	seen := map[string]struct{}{}

	for i, backend := range chain.backends {
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
