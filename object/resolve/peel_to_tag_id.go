package resolve

import objectid "codeberg.org/lindenii/furgit/object/id"

// PeelToTagID returns id unchanged.
func (r *Resolver) PeelToTagID(id objectid.ObjectID) (objectid.ObjectID, error) {
	return id, nil
}
