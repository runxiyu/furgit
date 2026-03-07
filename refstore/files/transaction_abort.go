package files

import "fmt"

func (tx *Transaction) Abort() error {
	err := tx.ensureOpen()
	if err != nil {
		return err
	}

	tx.closed = true

	return nil
}

func (tx *Transaction) ensureOpen() error {
	if tx.closed {
		return fmt.Errorf("refstore/files: transaction already closed")
	}

	return nil
}
