package stored

import objectid "lindenii.org/go/furgit/object/id"

// ID returns the object ID.
func (stored *Stored[T]) ID() objectid.ObjectID {
	return stored.id
}
