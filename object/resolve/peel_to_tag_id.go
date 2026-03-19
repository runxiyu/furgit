package resolve

import "codeberg.org/lindenii/furgit/objectid"

// PeelToTagID returns id unchanged.
func (r *Resolver) PeelToTagID(id objectid.ObjectID) (objectid.ObjectID, error) {
	return id, nil
}
