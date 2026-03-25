package resolve

import (
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
)

// PeelToTag returns the tag at id without further peeling.
func (r *Resolver) PeelToTag(id objectid.ObjectID) (*stored.Stored[*object.Tag], error) {
	return r.ExactTag(id)
}
