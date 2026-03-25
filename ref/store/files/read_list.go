package files

import (
	"errors"
	"path"
	"slices"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/ref/store"
)

// List lists references from the visible files ref namespace.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	matchAll := pattern == ""
	if !matchAll {
		_, err := path.Match(pattern, "HEAD")
		if err != nil {
			return nil, err
		}
	}

	looseNames, err := store.collectLooseRefNames()
	if err != nil {
		return nil, err
	}

	packed, err := store.readPackedRefs()
	if err != nil {
		return nil, err
	}

	byName := make(map[string]ref.Ref, len(looseNames)+len(packed.byName))
	for _, detached := range packed.ordered {
		byName[detached.Name()] = detached
	}

	for _, name := range looseNames {
		resolved, resolveErr := store.readLooseRef(name)
		if resolveErr != nil {
			if errors.Is(resolveErr, refstore.ErrReferenceNotFound) {
				delete(byName, name)

				continue
			}

			return nil, resolveErr
		}

		byName[name] = resolved
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		if !matchAll {
			matched, matchErr := path.Match(pattern, name)
			if matchErr != nil {
				return nil, matchErr
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
		refs = append(refs, byName[name])
	}

	return refs, nil
}
