package refstore

import "codeberg.org/lindenii/furgit/objectid"

// Batch applies reference operations immediately, one operation at a time.
//
// Unlike Transaction, Batch does not stage changes for one atomic commit.
// Each method attempts its update immediately and returns that operation's
// error, if any.
type Batch interface {
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

	// Close releases any resources associated with the batch.
	Close() error
}
