package files

import "slices"

func (executor *refUpdateExecutor) prepareUpdateLocks(prepared []preparedUpdate) error {
	lockNames := make([]string, 0, len(prepared))
	for _, item := range prepared {
		lockNames = append(lockNames, updateTargetKey(item.target.loc))
	}

	slices.Sort(lockNames)

	for _, lockKey := range lockNames {
		lockPath := refPathFromKey(lockKey)
		err := executor.createUpdateLock(lockPath)
		if err != nil {
			for _, item := range prepared {
				if updateTargetKey(item.target.loc) == lockKey {
					return wrapUpdateError(item.op.name, err)
				}
			}

			return err
		}
	}

	return nil
}
