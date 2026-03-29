package stored

import objectid "codeberg.org/lindenii/furgit/object/id"

// ID returns the object ID.
func (stored *Stored[T]) ID() objectid.ObjectID {
	return stored.id
}
