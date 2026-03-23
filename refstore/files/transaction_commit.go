package files

func (tx *Transaction) Commit() error {
	executor := &refUpdateExecutor{store: tx.store}
	prepared, err := executor.prepareUpdates(tx.ops)
	if err != nil {
		return err
	}

	return executor.commitPreparedUpdates(prepared)
}
