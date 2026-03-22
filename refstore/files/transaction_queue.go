package files

func (tx *Transaction) queue(op txOp) error {
	err := tx.validateOp(op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}
