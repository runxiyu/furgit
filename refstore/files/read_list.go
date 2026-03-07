package files

import (
	"errors"
	"os"
	"path"
	"slices"
	"strings"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
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

func (store *Store) collectLooseRefNames() ([]string, error) {
	names := make([]string, 0, 16)
	seen := make(map[string]struct{}, 16)

	_, err := store.gitRoot.Stat("HEAD")
	if err == nil {
		names = append(names, "HEAD")
		seen["HEAD"] = struct{}{}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var walk func(*os.Root, string) error

	walk = func(root *os.Root, dir string) error {
		file, openErr := root.Open(dir)
		if openErr != nil {
			if errors.Is(openErr, os.ErrNotExist) {
				return nil
			}

			return openErr
		}

		defer func() { _ = file.Close() }()

		entries, readErr := file.ReadDir(-1)
		if readErr != nil {
			return readErr
		}

		for _, entry := range entries {
			name := path.Join(dir, entry.Name())
			if entry.IsDir() {
				err := walk(root, name)
				if err != nil {
					return err
				}

				continue
			}

			if strings.HasSuffix(name, ".lock") {
				continue
			}

			if _, ok := seen[name]; ok {
				continue
			}

			seen[name] = struct{}{}
			names = append(names, name)
		}

		return nil
	}

	err = walk(store.commonRoot, "refs")
	if err != nil {
		return nil, err
	}

	err = walk(store.gitRoot, "refs")
	if err != nil {
		return nil, err
	}

	return names, nil
}
