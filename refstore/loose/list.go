package loose

import (
	"errors"
	"os"
	"path"
	"slices"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// List lists loose references matching pattern.
//
// Pattern uses path.Match syntax against full reference names.
// Empty pattern matches all references.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	matchAll := pattern == ""
	if !matchAll {
		_, err := path.Match(pattern, "HEAD")
		if err != nil {
			return nil, err
		}
	}

	names, err := store.collectLooseRefNames()
	if err != nil {
		return nil, err
	}

	slices.Sort(names)

	refs := make([]ref.Ref, 0, len(names))
	for _, name := range names {
		if !matchAll {
			matched, err := path.Match(pattern, name)
			if err != nil {
				return nil, err
			}

			if !matched {
				continue
			}
		}

		resolved, err := store.resolveOne(name)
		if err != nil {
			if errors.Is(err, refstore.ErrReferenceNotFound) {
				continue
			}

			return nil, err
		}

		refs = append(refs, resolved)
	}

	return refs, nil
}

// collectLooseRefNames returns loose ref names available in this backend.
func (store *Store) collectLooseRefNames() ([]string, error) {
	names := make([]string, 0, 16)

	_, err := store.root.Stat("HEAD")
	if err == nil {
		names = append(names, "HEAD")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var walk func(string) error

	walk = func(dir string) error {
		file, err := store.root.Open(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}

			return err
		}

		defer func() { _ = file.Close() }()

		entries, err := file.ReadDir(-1)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			name := path.Join(dir, entry.Name())
			if entry.IsDir() {
				err := walk(name)
				if err != nil {
					return err
				}

				continue
			}

			names = append(names, name)
		}

		return nil
	}

	err = walk("refs")
	if err != nil {
		return nil, err
	}

	return names, nil
}
