package memory

import (
	objectid "lindenii.org/go/furgit/object/id"
	refstore "lindenii.org/go/furgit/ref/store"
)

// Transaction stages in-memory updates for one atomic commit.
type Transaction struct {
	store *Store
	ops   []queuedUpdate
}

var _ refstore.Transaction = (*Transaction)(nil)

// BeginTransaction creates one new in-memory transaction.
//
//nolint:ireturn
func (store *Store) BeginTransaction() (refstore.Transaction, error) {
	return &Transaction{
		store: store,
		ops:   make([]queuedUpdate, 0, 8),
	}, nil
}

// Create queues a detached reference creation.
func (tx *Transaction) Create(name string, newID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

// Update queues a detached reference update.
func (tx *Transaction) Update(name string, newID, oldID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

// Delete queues a detached reference deletion.
func (tx *Transaction) Delete(name string, oldID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

// Verify queues a detached reference verification.
func (tx *Transaction) Verify(name string, oldID objectid.ObjectID) error {
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

	prepared, err := prepareUpdates(tx.store.refs, tx.ops)
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
	err := validateQueuedUpdate(op)
	if err != nil {
		return err
	}

	tx.ops = append(tx.ops, op)

	return nil
}
