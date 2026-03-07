package files

import (
	"errors"
	"os"
	"path"
)

func (tx *Transaction) tryRemoveEmptyParents(name string) {
	loc := tx.store.loosePath(name)
	tx.tryRemoveEmptyParentPaths(loc.root, loc.path)
}

func (tx *Transaction) tryRemoveEmptyParentPaths(kind rootKind, name string) {
	root := tx.store.rootFor(kind)
	dir := path.Dir(name)

	for dir != "." && dir != "/" {
		err := root.Remove(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return
			}

			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return
			}

			return
		}

		dir = path.Dir(dir)
	}
}
