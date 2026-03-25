package files

func (executor *refUpdateExecutor) commitPreparedUpdates(prepared []preparedUpdate) (err error) {
	defer func() {
		_ = executor.cleanupPreparedUpdates(prepared)
	}()

	for _, item := range prepared {
		if item.op.kind == updateDelete || item.op.kind == updateDeleteSymbolic || item.op.kind == updateVerify || item.op.kind == updateVerifySymbolic {
			continue
		}

		err = executor.writePreparedLooseUpdate(item)
		if err != nil {
			return wrapUpdateError(item.op.name, err)
		}
	}

	err = executor.applyPackedRefDeletes(prepared)
	if err != nil {
		return err
	}

	return executor.removeDeletedLooseRefs(prepared)
}
