package fetch

import objectid "codeberg.org/lindenii/furgit/object/id"

// Size returns the object content size at id.
//
// Labels: Life-Parent.
func (r *Fetcher) Size(id objectid.ObjectID) (int64, error) {
	return r.store.ReadSize(id)
}
