package resolve

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/object/tag"
)

// PeelToTag returns the tag at id without further peeling.
func (r *Resolver) PeelToTag(id objectid.ObjectID) (*stored.Stored[*tag.Tag], error) {
	return r.ExactTag(id)
}
