package resolve

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
)

// PeelToTag returns the tag at id without further peeling.
func (r *Resolver) PeelToTag(id objectid.ObjectID) (*stored.Stored[*object.Tag], error) {
	return r.ExactTag(id)
}
