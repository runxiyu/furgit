package ref

import objectid "lindenii.org/go/furgit/object/id"

// Detached points directly to an object ID.
//
// Labels: MT-Unsafe.
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
