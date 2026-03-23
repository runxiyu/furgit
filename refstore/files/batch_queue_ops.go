package files

import "codeberg.org/lindenii/furgit/objectid"

func (batch *Batch) Create(name string, newID objectid.ObjectID) {
	batch.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

func (batch *Batch) Update(name string, newID, oldID objectid.ObjectID) {
	batch.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

func (batch *Batch) Delete(name string, oldID objectid.ObjectID) {
	batch.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

func (batch *Batch) Verify(name string, oldID objectid.ObjectID) {
	batch.queue(queuedUpdate{name: name, kind: updateVerify, oldID: oldID})
}

func (batch *Batch) CreateSymbolic(name, newTarget string) {
	batch.queue(queuedUpdate{name: name, kind: updateCreateSymbolic, newTarget: newTarget})
}

func (batch *Batch) UpdateSymbolic(name, newTarget, oldTarget string) {
	batch.queue(queuedUpdate{name: name, kind: updateReplaceSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

func (batch *Batch) DeleteSymbolic(name, oldTarget string) {
	batch.queue(queuedUpdate{name: name, kind: updateDeleteSymbolic, oldTarget: oldTarget})
}

func (batch *Batch) VerifySymbolic(name, oldTarget string) {
	batch.queue(queuedUpdate{name: name, kind: updateVerifySymbolic, oldTarget: oldTarget})
}
