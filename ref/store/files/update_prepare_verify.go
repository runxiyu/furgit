package files

func (executor *refUpdateExecutor) verifyPreparedUpdates(prepared []preparedUpdate) error {
	for i := range prepared {
		item := &prepared[i]

		refState, err := executor.directRead(item.target.name)
		if err != nil {
			return wrapUpdateError(item.op.name, err)
		}

		item.target.ref = refState

		err = executor.verifyPreparedUpdateCurrent(*item)
		if err != nil {
			return err
		}
	}

	return nil
}
