package stored

import (
	"lindenii.org/go/furgit/object"
	objectid "lindenii.org/go/furgit/object/id"
)

// Stored represents a stored object,
// i.e., an object along with its object ID.
type Stored[T object.Object] struct {
	id  objectid.ObjectID
	obj T
}
