package files

import (
	"errors"
	"os"
	"path"
	"strings"
)

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
