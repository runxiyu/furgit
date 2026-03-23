package files

func (batch *Batch) queue(op queuedUpdate) {
	batch.ops = append(batch.ops, op)
}
