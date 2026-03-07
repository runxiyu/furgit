package files

import (
	"errors"
	"os"
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
