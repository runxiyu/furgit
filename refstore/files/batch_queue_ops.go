package files

import "codeberg.org/lindenii/furgit/objectid"

func (batch *Batch) Create(name string, newID objectid.ObjectID) {
	batch.queue(txOp{name: name, kind: txCreate, newID: newID})
}

func (batch *Batch) Update(name string, newID, oldID objectid.ObjectID) {
	batch.queue(txOp{name: name, kind: txUpdate, newID: newID, oldID: oldID})
}

func (batch *Batch) Delete(name string, oldID objectid.ObjectID) {
	batch.queue(txOp{name: name, kind: txDelete, oldID: oldID})
}

func (batch *Batch) Verify(name string, oldID objectid.ObjectID) {
	batch.queue(txOp{name: name, kind: txVerify, oldID: oldID})
}

func (batch *Batch) CreateSymbolic(name, newTarget string) {
	batch.queue(txOp{name: name, kind: txCreateSymbolic, newTarget: newTarget})
}

func (batch *Batch) UpdateSymbolic(name, newTarget, oldTarget string) {
	batch.queue(txOp{name: name, kind: txUpdateSymbolic, newTarget: newTarget, oldTarget: oldTarget})
}

func (batch *Batch) DeleteSymbolic(name, oldTarget string) {
	batch.queue(txOp{name: name, kind: txDeleteSymbolic, oldTarget: oldTarget})
}

func (batch *Batch) VerifySymbolic(name, oldTarget string) {
	batch.queue(txOp{name: name, kind: txVerifySymbolic, oldTarget: oldTarget})
}
