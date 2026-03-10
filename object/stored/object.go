package stored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// StoredObject is a parsed object paired with its storage ID.
type StoredObject interface {
	// ID returns the object ID the object was loaded from.
	ID() objectid.ObjectID
	// Object returns the parsed object value.
	Object() object.Object
}
