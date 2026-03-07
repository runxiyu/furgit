package files

func (batch *Batch) queue(op txOp) {
	if batch.closed {
		return
	}

	batch.ops = append(batch.ops, op)
}
