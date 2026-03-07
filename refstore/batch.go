package refstore

import "codeberg.org/lindenii/furgit/objectid"

// Batch stages reference operations for one non-atomic apply.
//
// Unlike Transaction, Batch may reject some queued operations while still
// applying others successfully when Apply runs.
type Batch interface {
	// Create creates one detached reference, requiring that the logical
	// reference does not already exist.
	Create(name string, newID objectid.ObjectID)
	// Update updates one detached reference, requiring that the current logical
	// reference value matches oldID.
	Update(name string, newID, oldID objectid.ObjectID)
	// Delete deletes one detached reference, requiring that the current logical
	// reference value matches oldID.
	Delete(name string, oldID objectid.ObjectID)
	// Verify verifies that the current logical reference value matches oldID.
	Verify(name string, oldID objectid.ObjectID)

	// CreateSymbolic creates one symbolic reference, requiring that the named
	// reference does not already exist.
	CreateSymbolic(name, newTarget string)
	// UpdateSymbolic updates one symbolic reference directly, requiring that its
	// current target matches oldTarget.
	UpdateSymbolic(name, newTarget, oldTarget string)
	// DeleteSymbolic deletes one symbolic reference directly, requiring that its
	// current target matches oldTarget.
	DeleteSymbolic(name, oldTarget string)
	// VerifySymbolic verifies that the named symbolic reference currently points
	// at oldTarget.
	VerifySymbolic(name, oldTarget string)

	// Apply validates and applies queued operations, returning one result per
	// queued operation in order. Fatal backend failures are returned separately.
	Apply() ([]BatchResult, error)
	// Abort abandons the batch and releases any resources it holds.
	Abort() error
}

// BatchResult reports the outcome for one queued batch operation.
type BatchResult struct {
	Name  string
	Error error
}
