package fetch

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// Header returns the object type and content size at id.
//
// Labels: Life-Parent.
func (r *Fetcher) Header(id objectid.ObjectID) (objecttype.Type, int64, error) {
	ty, size, err := r.store.ReadHeader(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, wrapObjectReadError(id, err)
	}

	return ty, size, nil
}
