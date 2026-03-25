package files

import (
	"errors"
	"os"
	"path"
)

func (executor *refUpdateExecutor) tryRemoveEmptyParents(name string) {
	loc := executor.store.loosePath(name)
	executor.tryRemoveEmptyParentPaths(loc.root, loc.path)
}

func (executor *refUpdateExecutor) tryRemoveEmptyParentPaths(kind rootKind, name string) {
	root := executor.store.rootFor(kind)
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
