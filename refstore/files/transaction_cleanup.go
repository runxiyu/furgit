package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
)

func (tx *Transaction) cleanup(prepared []preparedTxOp) error {
	var firstErr error

	lockNames := make([]string, 0, len(prepared)+1)
	for _, item := range prepared {
		lockNames = append(lockNames, tx.targetKey(item.target.loc))
	}

	lockNames = append(lockNames, tx.targetKey(refPath{root: rootCommon, path: "packed-refs"}))
	slices.Sort(lockNames)
	lockNames = slices.Compact(lockNames)

	for _, lockKey := range lockNames {
		lockPath := refPathFromKey(lockKey)
		lockName := lockPath.path + ".lock"
		root := tx.store.rootFor(lockPath.root)

		err := root.Remove(lockName)
		if err == nil || errors.Is(err, os.ErrNotExist) {
			tx.tryRemoveEmptyParentPaths(lockPath.root, lockName)

			continue
		}

		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

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

func (tx *Transaction) removeEmptyDirTree(name refPath) error {
	root := tx.store.rootFor(name.root)

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

	return tx.removeEmptyDirTreeRecursive(name)
}

func (tx *Transaction) removeEmptyDirTreeRecursive(name refPath) error {
	root := tx.store.rootFor(name.root)

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

		err = tx.removeEmptyDirTreeRecursive(refPath{
			root: name.root,
			path: path.Join(name.path, entry.Name()),
		})
		if err != nil {
			return err
		}
	}

	return root.Remove(name.path)
}
