package fetch

import objectid "codeberg.org/lindenii/furgit/object/id"

// PeelToTagID returns id unchanged.
func (r *Fetcher) PeelToTagID(id objectid.ObjectID) (objectid.ObjectID, error) {
	return id, nil
}
