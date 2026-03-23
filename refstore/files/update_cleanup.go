package files

import (
	"errors"
	"os"
	"slices"
)

func (executor *refUpdateExecutor) cleanupPreparedUpdates(prepared []preparedUpdate) error {
	var firstErr error

	lockNames := make([]string, 0, len(prepared)+1)
	for _, item := range prepared {
		lockNames = append(lockNames, updateTargetKey(item.target.loc))
	}

	lockNames = append(lockNames, updateTargetKey(refPath{root: rootCommon, path: "packed-refs"}))
	slices.Sort(lockNames)
	lockNames = slices.Compact(lockNames)

	for _, lockKey := range lockNames {
		lockPath := refPathFromKey(lockKey)
		lockName := lockPath.path + ".lock"
		root := executor.store.rootFor(lockPath.root)

		err := root.Remove(lockName)
		if err == nil || errors.Is(err, os.ErrNotExist) {
			executor.tryRemoveEmptyParentPaths(lockPath.root, lockName)

			continue
		}

		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
