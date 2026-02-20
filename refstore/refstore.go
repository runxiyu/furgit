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
	Resolve(name string) (ref.Ref, error)
	ResolveFully(name string) (ref.Detached, error)
	List(pattern string) ([]ref.Ref, error)
	Close() error
}
