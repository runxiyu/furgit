// Package ref provides Git reference models.
package ref

import "codeberg.org/lindenii/furgit/objectid"

// Ref is a Git reference.
//
// Implementations must be in this package.
type Ref interface {
	isRef()
	Name() string
}

// Detached points directly to an object ID.
type Detached struct {
	RefName string
	ID      objectid.ObjectID

	// Peeled is the peeled target when available (for annotated tags).
	//
	// This field is optional backend-provided metadata. Backends that do not
	// have peel metadata available may leave it nil.
	Peeled *objectid.ObjectID
}

// Name returns the fully-qualified reference name.
func (ref Detached) Name() string {
	return ref.RefName
}

func (Detached) isRef() {}

// Symbolic points to another reference name.
type Symbolic struct {
	RefName string
	Target  string
}

// Name returns the fully-qualified reference name.
func (ref Symbolic) Name() string {
	return ref.RefName
}

func (Symbolic) isRef() {}
