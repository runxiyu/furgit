package files

import objectid "lindenii.org/go/furgit/object/id"

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
