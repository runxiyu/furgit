package files

import "errors"

func (batch *Batch) Abort() error {
	if batch.closed {
		return errors.New("refstore/files: batch already closed")
	}

	batch.closed = true

	return nil
}
