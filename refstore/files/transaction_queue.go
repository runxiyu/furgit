package files

func (tx *Transaction) queue(op txOp) error {
	err := tx.ensureOpen()
	if err != nil {
		return err
	}

	err = tx.validateOp(op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}
