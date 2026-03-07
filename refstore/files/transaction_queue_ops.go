package files

import "codeberg.org/lindenii/furgit/objectid"

func (tx *Transaction) Create(name string, newID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txCreate, newID: newID})
}

func (tx *Transaction) Update(name string, newID, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txUpdate, newID: newID, oldID: oldID})
}

func (tx *Transaction) Delete(name string, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txDelete, oldID: oldID})
}

func (tx *Transaction) Verify(name string, oldID objectid.ObjectID) error {
	return tx.queue(txOp{name: name, kind: txVerify, oldID: oldID})
}

func (tx *Transaction) CreateSymbolic(name, newTarget string) error {
	return tx.queue(txOp{name: name, kind: txCreateSymbolic, newTarget: newTarget})
}

func (tx *Transaction) UpdateSymbolic(name, newTarget, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txUpdateSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

func (tx *Transaction) DeleteSymbolic(name, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txDeleteSymbolic, oldTarget: oldTarget})
}

func (tx *Transaction) VerifySymbolic(name, oldTarget string) error {
	return tx.queue(txOp{name: name, kind: txVerifySymbolic, oldTarget: oldTarget})
}
