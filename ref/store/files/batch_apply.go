package files

import refstore "codeberg.org/lindenii/furgit/ref/store"

// Apply validates and applies the queued updates.
func (batch *Batch) Apply() ([]refstore.BatchResult, error) {
	results := make([]refstore.BatchResult, len(batch.ops))
	remainingIdx := make([]int, 0, len(batch.ops))
	remainingOps := make([]queuedUpdate, 0, len(batch.ops))
	seenTargets := make(map[string]struct{}, len(batch.ops))
	executor := &refUpdateExecutor{store: batch.store}

	for i, op := range batch.ops {
		results[i].Name = op.name

		target, err := executor.resolveQueuedUpdateTarget(op)
		if err != nil {
			if isBatchRejected(err) {
				results[i].Status = refstore.BatchStatusRejected
				results[i].Error = batchResultError(err)

				continue
			}

			results[i].Status = refstore.BatchStatusFatal
			results[i].Error = batchResultError(err)

			for j := i + 1; j < len(results); j++ {
				results[j].Name = batch.ops[j].name
				results[j].Status = refstore.BatchStatusNotAttempted
				results[j].Error = batchResultError(err)
			}

			return results, err
		}

		targetKey := updateTargetKey(target.loc)
		if _, exists := seenTargets[targetKey]; exists {
			results[i].Status = refstore.BatchStatusRejected
			results[i].Error = &refstore.DuplicateUpdateError{}

			continue
		}

		seenTargets[targetKey] = struct{}{}

		remainingIdx = append(remainingIdx, i)
		remainingOps = append(remainingOps, op)
	}

	for len(remainingOps) > 0 {
		prepared, err := executor.prepareUpdates(remainingOps)
		if err == nil {
			err = executor.commitPreparedUpdates(prepared)
			if err == nil {
				for _, idx := range remainingIdx {
					results[idx].Status = refstore.BatchStatusApplied
				}

				return results, nil
			}

			fatalName := batchResultName(err)

			fatalMarked := false
			for i, idx := range remainingIdx {
				if !fatalMarked && remainingOps[i].name == fatalName && fatalName != "" {
					results[idx].Status = refstore.BatchStatusFatal
					results[idx].Error = batchResultError(err)
					fatalMarked = true

					continue
				}

				results[idx].Status = refstore.BatchStatusNotAttempted
				results[idx].Error = batchResultError(err)
			}

			return results, err
		}

		if !isBatchRejected(err) {
			fatalName := batchResultName(err)

			fatalMarked := false
			for i, idx := range remainingIdx {
				if !fatalMarked && remainingOps[i].name == fatalName && fatalName != "" {
					results[idx].Status = refstore.BatchStatusFatal
					results[idx].Error = batchResultError(err)
					fatalMarked = true

					continue
				}

				results[idx].Status = refstore.BatchStatusNotAttempted
				results[idx].Error = batchResultError(err)
			}

			return results, err
		}

		name := batchResultName(err)
		rejectedAt := -1

		for i, op := range remainingOps {
			if op.name == name {
				rejectedAt = i

				break
			}
		}

		if rejectedAt < 0 {
			for _, idx := range remainingIdx {
				results[idx].Status = refstore.BatchStatusNotAttempted
				results[idx].Error = batchResultError(err)
			}

			return results, err
		}

		results[remainingIdx[rejectedAt]].Status = refstore.BatchStatusRejected
		results[remainingIdx[rejectedAt]].Error = batchResultError(err)
		remainingIdx = append(remainingIdx[:rejectedAt], remainingIdx[rejectedAt+1:]...)
		remainingOps = append(remainingOps[:rejectedAt], remainingOps[rejectedAt+1:]...)
	}

	return results, nil
}
