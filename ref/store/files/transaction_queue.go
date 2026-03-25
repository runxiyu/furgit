package files

func (tx *Transaction) queue(op queuedUpdate) error {
	err := (&refUpdateExecutor{store: tx.store}).validateQueuedUpdate(op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}
