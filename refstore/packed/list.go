package packed

import (
	"path"

	"codeberg.org/lindenii/furgit/ref"
)

// List lists packed references matching pattern.
//
// Pattern uses path.Match syntax against full reference names.
// Empty pattern matches all references.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	matchAll := pattern == ""
	if !matchAll {
		_, err := path.Match(pattern, "refs/heads/main")
		if err != nil {
			return nil, err
		}
	}

	refs := make([]ref.Ref, 0, len(store.ordered))
	for _, entry := range store.ordered {
		if !matchAll {
			matched, err := path.Match(pattern, entry.Name())
			if err != nil {
				return nil, err
			}

			if !matched {
				continue
			}
		}

		refs = append(refs, entry)
	}

	return refs, nil
}
