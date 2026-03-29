package stored

import (
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Stored represents a stored object,
// i.e., an object along with its object ID.
type Stored[T object.Object] struct {
	id  objectid.ObjectID
	obj T
}
