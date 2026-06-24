package memory

import (
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store"
)

// Transaction stages in-memory updates for one atomic commit.
type Transaction struct {
	store *Memory
	ops   []queuedUpdate
}

var _ store.Transaction = (*Transaction)(nil)

// BeginTransaction creates one new in-memory transaction.
func (memory *Memory) BeginTransaction() (store.Transaction, error) {
	return &Transaction{
		store: memory,
		ops:   make([]queuedUpdate, 0, 8),
	}, nil
}

// Create queues a direct reference creation.
func (tx *Transaction) Create(name string, newID id.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

// Update queues a direct reference update.
func (tx *Transaction) Update(name string, newID, oldID id.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

// Delete queues a direct reference deletion.
func (tx *Transaction) Delete(name string, oldID id.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

// Verify queues a direct reference verification.
func (tx *Transaction) Verify(name string, oldID id.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateVerify, oldID: oldID})
}

// CreateSymbolic queues a symbolic reference creation.
func (tx *Transaction) CreateSymbolic(name, newTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateCreateSymbolic, newTarget: newTarget})
}

// UpdateSymbolic queues a symbolic reference update.
func (tx *Transaction) UpdateSymbolic(name, newTarget, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateReplaceSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

// DeleteSymbolic queues a symbolic reference deletion.
func (tx *Transaction) DeleteSymbolic(name, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateDeleteSymbolic, oldTarget: oldTarget})
}

// VerifySymbolic queues a symbolic reference verification.
func (tx *Transaction) VerifySymbolic(name, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateVerifySymbolic, oldTarget: oldTarget})
}

// Commit validates and applies the queued updates atomically.
func (tx *Transaction) Commit() error {
	tx.store.mu.Lock()
	defer tx.store.mu.Unlock()

	prepared, _, err := prepareUpdates(tx.store.refs, tx.ops)
	if err != nil {
		return err
	}

	next := cloneRefs(tx.store.refs)
	applyPreparedUpdates(next, prepared)
	tx.store.refs = next

	return nil
}

// Abort abandons the transaction.
func (tx *Transaction) Abort() error {
	return nil
}

func (tx *Transaction) queue(op queuedUpdate) error {
	err := validateQueuedUpdate(tx.store.objectFormat, op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}
