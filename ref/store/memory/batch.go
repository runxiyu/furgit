package memory

import (
	objectid "lindenii.org/go/furgit/object/id"
	refstore "lindenii.org/go/furgit/ref/store"
)

// Batch stages in-memory updates for one subset commit.
type Batch struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Batch = (*Batch)(nil)

// BeginBatch creates one new in-memory batch.
//
//nolint:ireturn
func (store *Store) BeginBatch() (refstore.Batch, error) {
	return &Batch{
		store: store,
		ops:   make([]queuedUpdate, 0, 8),
	}, nil
}

// Create queues a detached reference creation.
func (batch *Batch) Create(name string, newID objectid.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

// Update queues a detached reference update.
func (batch *Batch) Update(name string, newID, oldID objectid.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

// Delete queues a detached reference deletion.
func (batch *Batch) Delete(name string, oldID objectid.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

// Verify queues a detached reference verification.
func (batch *Batch) Verify(name string, oldID objectid.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateVerify, oldID: oldID})
}

// CreateSymbolic queues a symbolic reference creation.
func (batch *Batch) CreateSymbolic(name, newTarget string) error {
	return batch.queue(queuedUpdate{name: name, kind: updateCreateSymbolic, newTarget: newTarget})
}

// UpdateSymbolic queues a symbolic reference update.
func (batch *Batch) UpdateSymbolic(name, newTarget, oldTarget string) error {
	return batch.queue(queuedUpdate{name: name, kind: updateReplaceSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

// DeleteSymbolic queues a symbolic reference deletion.
func (batch *Batch) DeleteSymbolic(name, oldTarget string) error {
	return batch.queue(queuedUpdate{name: name, kind: updateDeleteSymbolic, oldTarget: oldTarget})
}

// VerifySymbolic queues a symbolic reference verification.
func (batch *Batch) VerifySymbolic(name, oldTarget string) error {
	return batch.queue(queuedUpdate{name: name, kind: updateVerifySymbolic, oldTarget: oldTarget})
}

// Apply validates queued operations,
// drops rejected operations,
// and applies the remaining compatible set.
// Concurrent readers observe
// either the pre-Apply state
// or the post-Apply state.
func (batch *Batch) Apply() ([]refstore.BatchResult, error) {
	results := make([]refstore.BatchResult, len(batch.ops))
	remainingIdx := make([]int, 0, len(batch.ops))
	remainingOps := make([]queuedUpdate, 0, len(batch.ops))
	seenTargets := make(map[string]struct{}, len(batch.ops))

	batch.store.mu.Lock()
	defer batch.store.mu.Unlock()

	for i, op := range batch.ops {
		results[i].Name = op.name

		target, err := resolveQueuedUpdateTarget(batch.store.refs, op)
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

		if _, exists := seenTargets[target.name]; exists {
			results[i].Status = refstore.BatchStatusRejected
			results[i].Error = &refstore.DuplicateUpdateError{}

			continue
		}

		seenTargets[target.name] = struct{}{}

		remainingIdx = append(remainingIdx, i)
		remainingOps = append(remainingOps, op)
	}

	for len(remainingOps) > 0 {
		prepared, err := prepareUpdates(batch.store.refs, remainingOps)
		if err == nil {
			next := cloneRefs(batch.store.refs)
			applyPreparedUpdates(next, prepared)
			batch.store.refs = next

			for _, idx := range remainingIdx {
				results[idx].Status = refstore.BatchStatusApplied
			}

			return results, nil
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

// Abort abandons the batch.
func (batch *Batch) Abort() error {
	return nil
}

func (batch *Batch) queue(op queuedUpdate) error {
	err := validateQueuedUpdate(op)
	if err != nil {
		return err
	}

	batch.ops = append(batch.ops, op)

	return nil
}
