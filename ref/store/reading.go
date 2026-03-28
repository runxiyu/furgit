package refstore

import "codeberg.org/lindenii/furgit/ref"

// ReadingStore reads Git references.
//
// Labels: MT-Safe.
type ReadingStore interface {
	// Resolve resolves a reference name to either a symbolic or detached ref.
	//
	// Implementations should return value forms ([ref.Detached] or [ref.Symbolic]),
	// not pointer forms. If the reference does not exist, implementations should
	// return [ErrReferenceNotFound].
	//
	// Labels: Life-Parent.
	Resolve(name string) (ref.Ref, error)
	// ResolveToDetached resolves a reference name to a detached object ID.
	//
	// Implementations may use backend-local lookup semantics for symbolic hops.
	// Callers that need cross-backend symbolic resolution (for example in a
	// chain of stores) should prefer repeatedly calling Resolve.
	//
	// ResolveToDetached resolves symbolic references only. It does not imply peeling
	// annotated tag objects.
	//
	// Labels: Life-Parent.
	ResolveToDetached(name string) (ref.Detached, error)
	// List returns references matching pattern.
	//
	// The exact pattern language is backend-defined.
	//
	// Labels: Life-Parent.
	List(pattern string) ([]ref.Ref, error)
	// Close releases resources associated with the store.
	//
	// Transactions and batches borrowing the store are invalid after Close.
	//
	// Labels: Close-UB.
	Close() error
}
