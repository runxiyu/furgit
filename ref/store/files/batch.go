package files

import (
	"errors"

	"lindenii.org/go/furgit/object/id"
	refname "lindenii.org/go/furgit/ref/name"
	"lindenii.org/go/furgit/ref/store"
)

// Batch stages files-store updates for one non-atomic apply.
type Batch struct {
	files *Files
	ops   []queuedUpdate
}

var _ store.Batch = (*Batch)(nil)

// BeginBatch creates one new files batch.
func (files *Files) BeginBatch() (store.Batch, error) {
	return &Batch{
		files: files,
		ops:   make([]queuedUpdate, 0, 8),
	}, nil
}

// Create queues a direct reference creation.
func (batch *Batch) Create(name string, newID id.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

// Update queues a direct reference update.
func (batch *Batch) Update(name string, newID, oldID id.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

// Delete queues a direct reference deletion.
func (batch *Batch) Delete(name string, oldID id.ObjectID) error {
	return batch.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

// Verify queues a direct reference verification.
func (batch *Batch) Verify(name string, oldID id.ObjectID) error {
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
func (batch *Batch) Apply() ([]store.BatchResult, error) {
	results := make([]store.BatchResult, len(batch.ops))
	remainingIdx := make([]int, 0, len(batch.ops))
	remainingOps := make([]queuedUpdate, 0, len(batch.ops))
	seenTargets := make(map[refPath]struct{}, len(batch.ops))
	executor := &updateExecutor{files: batch.files}

	for i, op := range batch.ops {
		results[i].Name = op.name

		target, err := executor.resolveQueuedUpdateTarget(op)
		if err != nil {
			if isBatchRejected(err) {
				results[i].Status = store.BatchStatusRejected
				results[i].Error = err

				continue
			}

			markFatal(results, batch.ops, i, err)

			return results, err
		}

		if _, exists := seenTargets[target.loc]; exists {
			results[i].Status = store.BatchStatusRejected
			results[i].Error = store.ErrDuplicateUpdate

			continue
		}

		seenTargets[target.loc] = struct{}{}

		remainingIdx = append(remainingIdx, i)
		remainingOps = append(remainingOps, op)
	}

	return batch.applyRemaining(results, remainingIdx, remainingOps)
}

// Abort abandons the batch.
func (batch *Batch) Abort() error {
	return nil
}

// applyRemaining repeatedly prepares the remaining operations,
// dropping one rejected operation per round,
// until either the whole set applies cleanly or a fatal failure occurs.
func (batch *Batch) applyRemaining(results []store.BatchResult, remainingIdx []int, remainingOps []queuedUpdate) ([]store.BatchResult, error) {
	for len(remainingOps) > 0 {
		executor := &updateExecutor{files: batch.files}

		prepared, failedName, err := executor.prepare(remainingOps)
		if err == nil {
			err = executor.commit(prepared)
			if err != nil {
				markFatalRemaining(results, remainingIdx, remainingOps, "", err)

				return results, err
			}

			for _, idx := range remainingIdx {
				results[idx].Status = store.BatchStatusApplied
			}

			return results, nil
		}

		executor.cleanup()

		if !isBatchRejected(err) {
			markFatalRemaining(results, remainingIdx, remainingOps, failedName, err)

			return results, err
		}

		rejectedAt := indexOfName(remainingOps, failedName)
		if rejectedAt < 0 {
			for _, idx := range remainingIdx {
				results[idx].Status = store.BatchStatusNotAttempted
				results[idx].Error = err
			}

			return results, err
		}

		results[remainingIdx[rejectedAt]].Status = store.BatchStatusRejected
		results[remainingIdx[rejectedAt]].Error = err
		remainingIdx = append(remainingIdx[:rejectedAt], remainingIdx[rejectedAt+1:]...)
		remainingOps = append(remainingOps[:rejectedAt], remainingOps[rejectedAt+1:]...)
	}

	return results, nil
}

func (batch *Batch) queue(op queuedUpdate) error {
	err := validateQueuedUpdate(batch.files.objectFormat, op)
	if err != nil {
		return err
	}

	batch.ops = append(batch.ops, op)

	return nil
}

// isBatchRejected reports whether err is a per-operation rejection
// that should drop only the offending operation,
// rather than a fatal failure that aborts the whole batch.
func isBatchRejected(err error) bool {
	switch {
	case errors.Is(err, store.ErrReferenceNotFound),
		errors.Is(err, store.ErrCreateExists),
		errors.Is(err, store.ErrDuplicateUpdate),
		errors.Is(err, store.ErrExpectedDirect),
		errors.Is(err, store.ErrExpectedSymbolic),
		errors.Is(err, store.ErrInvalidValue),
		errors.Is(err, store.ErrSymbolicCycle),
		errors.Is(err, errBrokenRef),
		errors.Is(err, refname.ErrInvalidName):
		return true
	}

	if _, ok := errors.AsType[*store.NameConflictError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*store.WrongOldIDError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*store.WrongOldTargetError](err); ok {
		return true
	}

	return false
}

func markFatal(results []store.BatchResult, ops []queuedUpdate, at int, err error) {
	results[at].Status = store.BatchStatusFatal
	results[at].Error = err

	for j := at + 1; j < len(results); j++ {
		results[j].Name = ops[j].name
		results[j].Status = store.BatchStatusNotAttempted
		results[j].Error = err
	}
}

func markFatalRemaining(results []store.BatchResult, remainingIdx []int, remainingOps []queuedUpdate, failedName string, err error) {
	fatalMarked := false

	for i, idx := range remainingIdx {
		if !fatalMarked && failedName != "" && remainingOps[i].name == failedName {
			results[idx].Status = store.BatchStatusFatal
			results[idx].Error = err
			fatalMarked = true

			continue
		}

		results[idx].Status = store.BatchStatusNotAttempted
		results[idx].Error = err
	}
}

func indexOfName(ops []queuedUpdate, name string) int {
	for i, op := range ops {
		if op.name == name {
			return i
		}
	}

	return -1
}
