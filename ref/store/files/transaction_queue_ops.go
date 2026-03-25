package files

import objectid "codeberg.org/lindenii/furgit/object/id"

func (tx *Transaction) Create(name string, newID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateCreate, newID: newID})
}

func (tx *Transaction) Update(name string, newID, oldID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateReplace, newID: newID, oldID: oldID})
}

func (tx *Transaction) Delete(name string, oldID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateDelete, oldID: oldID})
}

func (tx *Transaction) Verify(name string, oldID objectid.ObjectID) error {
	return tx.queue(queuedUpdate{name: name, kind: updateVerify, oldID: oldID})
}

func (tx *Transaction) CreateSymbolic(name, newTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateCreateSymbolic, newTarget: newTarget})
}

func (tx *Transaction) UpdateSymbolic(name, newTarget, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateReplaceSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

func (tx *Transaction) DeleteSymbolic(name, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateDeleteSymbolic, oldTarget: oldTarget})
}

func (tx *Transaction) VerifySymbolic(name, oldTarget string) error {
	return tx.queue(queuedUpdate{name: name, kind: updateVerifySymbolic, oldTarget: oldTarget})
}
