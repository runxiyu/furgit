package files

import (
	"errors"

	"codeberg.org/lindenii/furgit/refstore"
)

func (batch *Batch) Apply() ([]refstore.BatchResult, error) {
	if batch.closed {
		return nil, errors.New("refstore/files: batch already closed")
	}

	results := make([]refstore.BatchResult, len(batch.ops))
	seen := make(map[string]struct{}, len(batch.ops))

	for i, op := range batch.ops {
		results[i].Name = op.name

		if _, exists := seen[op.name]; exists {
			batch.closed = true

			err := errors.New("refstore/files: duplicate batch operation for " + `"` + op.name + `"`)
			for j := i; j < len(results); j++ {
				results[j].Name = batch.ops[j].name
				results[j].Error = err
			}

			return results, err
		}

		seen[op.name] = struct{}{}
	}

	for i, op := range batch.ops {
		tx := &Transaction{
			store: batch.store,
			ops:   []txOp{op},
		}

		err := tx.validateOp(op)
		if err != nil {
			results[i].Error = err

			continue
		}

		err = tx.Commit()
		if err == nil {
			continue
		}

		if isBatchRejected(err) {
			results[i].Error = err

			continue
		}

		batch.closed = true
		results[i].Error = err

		for j := i + 1; j < len(results); j++ {
			results[j].Name = batch.ops[j].name
			results[j].Error = err
		}

		return results, err
	}

	batch.closed = true

	return results, nil
}
