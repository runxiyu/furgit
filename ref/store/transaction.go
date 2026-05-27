package refstore

import objectid "lindenii.org/go/furgit/object/id"

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
//
// Labels: MT-Unsafe.
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
	// Commit invalidates the receiver.
	Commit() error
	// Abort abandons the transaction and releases any resources it holds.
	//
	// Abort invalidates the receiver.
	Abort() error
}
