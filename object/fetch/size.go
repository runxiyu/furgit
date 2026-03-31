package fetch

import objectid "codeberg.org/lindenii/furgit/object/id"

// Size returns the object content size at id.
//
// Labels: Life-Parent.
func (r *Fetcher) Size(id objectid.ObjectID) (int64, error) {
	size, err := r.store.ReadSize(id)
	if err != nil {
		return 0, wrapObjectReadError(id, err)
	}

	return size, nil
}
