package files

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func (executor *refUpdateExecutor) removeEmptyDirTree(name refPath) error {
	root := executor.store.rootFor(name.root)

	info, err := root.Stat(name.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if !info.IsDir() {
		return nil
	}

	return executor.removeEmptyDirTreeRecursive(name)
}

func (executor *refUpdateExecutor) removeEmptyDirTreeRecursive(name refPath) error {
	root := executor.store.rootFor(name.root)

	dir, err := root.Open(name.path)
	if err != nil {
		return err
	}

	entries, err := dir.ReadDir(-1)
	_ = dir.Close()

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("refstore/files: non-empty directory blocks reference %q", name.path)
		}

		err = executor.removeEmptyDirTreeRecursive(refPath{
			root: name.root,
			path: path.Join(name.path, entry.Name()),
		})
		if err != nil {
			return err
		}
	}

	return root.Remove(name.path)
}
