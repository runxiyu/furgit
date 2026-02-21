// Package refstore provides interfaces for reference storage backends.
package refstore

import (
	"errors"

	"codeberg.org/lindenii/furgit/ref"
)

// ErrReferenceNotFound indicates that a reference does not exist in a backend.
// TODO: interface error? just like object not found in objectstore
var ErrReferenceNotFound = errors.New("refstore: reference not found")

// Store reads Git references.
type Store interface {
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
	// Shorten returns the shortest unambiguous shorthand for a full
	// reference name within this store's visible namespace.
	//
	// If name does not exist in this store, implementations should return
	// ErrReferenceNotFound.
	Shorten(name string) (string, error)
	// Close releases resources associated with the store.
	Close() error
}
