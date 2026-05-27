package files

import objectid "lindenii.org/go/furgit/object/id"

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
