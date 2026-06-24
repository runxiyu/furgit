package store

import "lindenii.org/go/furgit/ref"

// Reader reads Git references.
//
// Labels: MT-Safe.
type Reader interface {
	// Resolve resolves a reference name
	// to either a symbolic or direct ref.
	//
	// Implementations return value forms
	// ([ref.Direct] or [ref.Symbolic]),
	// not pointer forms.
	// If the reference does not exist,
	// implementations return [ErrReferenceNotFound].
	//
	// Labels: Life-Parent.
	Resolve(name string) (ref.Ref, error)

	// ResolveToDirect resolves a reference name to a direct reference,
	// following symbolic references until one is reached.
	//
	// It follows symbolic references only;
	// it does not peel annotated tag objects.
	//
	// Implementations may follow symbolic hops with backend-local lookup.
	// Callers that need cross-backend symbolic resolution
	// (for example across a chain of stores)
	// should prefer repeatedly calling Resolve.
	//
	// Labels: Life-Parent.
	ResolveToDirect(name string) (ref.Direct, error)
}
