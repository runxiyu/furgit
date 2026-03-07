// Package refstore provides interfaces for reference storage backends.
package refstore

import (
	"errors"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
)

// ErrReferenceNotFound indicates that a reference does not exist in a backend.
// TODO: Interface error? Just like object not found in objectstore.
var ErrReferenceNotFound = errors.New("refstore: reference not found")

// ReadingStore reads Git references.
type ReadingStore interface {
	// Resolve resolves a reference name to either a symbolic or detached ref.
	//
	// Implementations should return value forms (ref.Detached or ref.Symbolic),
	// not pointer forms.
	// If the reference does not exist, implementations should return
	// ErrReferenceNotFound.
	Resolve(name string) (ref.Ref, error)
	// ResolveFully resolves a reference name to a detached object ID.
	//
	// Implementations may use backend-local lookup semantics for symbolic hops.
	// Callers that need cross-backend symbolic resolution (for example in a
	// chain of stores) should prefer repeatedly calling Resolve.
	//
	// ResolveFully resolves symbolic references only. It does not imply peeling
	// annotated tag objects.
	ResolveFully(name string) (ref.Detached, error)
	// List returns references matching pattern.
	//
	// The exact pattern language is backend-defined.
	List(pattern string) ([]ref.Ref, error)
	// Close releases resources associated with the store.
	Close() error
}

// TransactionalStore begins atomic reference transactions.
//
// Not every readable reference store is writable. Implementations should only
// satisfy TransactionalStore when they can stage and commit reference updates
// atomically within that backend.
type TransactionalStore interface {
	// BeginTransaction creates one new mutable transaction.
	BeginTransaction() (Transaction, error)
}

// Transaction stages reference updates for one atomic commit.
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
	Commit() error
	// Abort abandons the transaction and releases any resources it holds.
	Abort() error
}
