package refstore

import objectid "codeberg.org/lindenii/furgit/object/id"

// Batch stages reference operations for one non-atomic apply.
//
// Unlike Transaction, Batch may reject some queued operations while still
// applying others successfully when Apply runs.
//
// A batch borrows its underlying store and is invalid after that store is
// closed.
//
// Labels: MT-Unsafe.
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

	// Apply validates and applies queued operations, returning one result per
	// queued operation in order. Fatal backend failures are returned separately.
	//
	// Malformed operations are rejected by the queueing methods above and do not
	// enter the batch.
	//
	// Apply invalidates the receiver.
	Apply() ([]BatchResult, error)
	// Abort abandons the batch and releases any resources it holds.
	//
	// Abort invalidates the receiver.
	Abort() error
}

// BatchStatus reports the outcome for one queued batch operation.
type BatchStatus uint8

const (
	BatchStatusApplied BatchStatus = iota
	BatchStatusRejected
	BatchStatusFatal
	BatchStatusNotAttempted
)

// BatchResult reports the outcome for one queued batch operation.
type BatchResult struct {
	Name   string
	Status BatchStatus
	Error  error
}
