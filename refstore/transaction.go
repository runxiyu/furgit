package refstore

import "codeberg.org/lindenii/furgit/objectid"

// Transaction stages reference updates for one atomic commit.
//
// A transaction borrows its underlying store and is invalid after that store
// is closed.
//
// Ordinary methods operate in dereference mode if name resolves to
// a symbolic ref, the operation applies to the final referent rather
// than to the symbolic ref itself.
//
// Symbolic methods operate on the named reference directly, without
// dereferencing symbolic refs.
type Transaction interface {
	// Create creates one detached reference, requiring that the logical
	// reference does not already exist.
	Create(name string, newID objectid.ObjectID) error
	// Update updates one detached reference, requiring that the current logical
	// reference value matches oldID.
	Update(name string, newID, oldID objectid.ObjectID) error
	// Delete deletes one detached reference, requiring that the current logical
	// reference value matches oldID.
	Delete(name string, oldID objectid.ObjectID) error
	// Verify verifies that the current logical reference value matches oldID.
	Verify(name string, oldID objectid.ObjectID) error

	// CreateSymbolic creates one symbolic reference, requiring that the named
	// reference does not already exist.
	CreateSymbolic(name, newTarget string) error
	// UpdateSymbolic updates one symbolic reference directly, requiring that its
	// current target matches oldTarget.
	UpdateSymbolic(name, newTarget, oldTarget string) error
	// DeleteSymbolic deletes one symbolic reference directly, requiring that its
	// current target matches oldTarget.
	DeleteSymbolic(name, oldTarget string) error
	// VerifySymbolic verifies that the named symbolic reference currently points
	// at oldTarget.
	VerifySymbolic(name, oldTarget string) error

	// Commit validates and applies all queued operations atomically.
	//
	// Commit is terminal. Further use of the transaction is undefined behavior.
	Commit() error
	// Abort abandons the transaction and releases any resources it holds.
	//
	// Abort is terminal. Further use of the transaction is undefined behavior.
	Abort() error
}
