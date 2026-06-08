package fetch

import oid "lindenii.org/go/furgit/object/id"

// Size returns the object content size at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) Size(id oid.ObjectID) (uint64, error) {
	size, err := fetcher.store.ReadSize(id)
	if err != nil {
		return 0, wrapObjectReadError(id, err)
	}

	return size, nil
}
