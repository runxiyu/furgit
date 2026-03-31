package files

func (batch *Batch) queue(op queuedUpdate) error {
	err := (&refUpdateExecutor{store: batch.store}).validateQueuedUpdate(op)
	if err != nil {
		return err
	}

	batch.ops = append(batch.ops, op)

	return nil
}
