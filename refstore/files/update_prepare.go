package files

func (executor *refUpdateExecutor) prepareUpdates(ops []queuedUpdate) (prepared []preparedUpdate, err error) {
	defer func() {
		if err != nil {
			_ = executor.cleanupPreparedUpdates(prepared)
		}
	}()

	prepared, err = executor.resolvePreparedUpdates(ops)
	if err != nil {
		return prepared, err
	}

	deleted, written := collectPreparedWrites(prepared)

	existing, err := executor.collectVisibleNames()
	if err != nil {
		return prepared, err
	}

	for _, name := range written {
		err = verifyRefnameAvailable(name, existing, written, deleted)
		if err != nil {
			return prepared, err
		}
	}

	err = executor.prepareUpdateLocks(prepared)
	if err != nil {
		return prepared, err
	}

	hasDeletes := len(deleted) > 0
	if hasDeletes {
		err = executor.createPackedRefsLock(executor.store.packedRefsTimeout)
		if err != nil {
			return prepared, err
		}
	}

	err = executor.verifyPreparedUpdates(prepared)
	if err != nil {
		return prepared, err
	}

	return prepared, nil
}
