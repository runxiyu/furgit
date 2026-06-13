package fetch

import (
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// Header returns the object type and content size at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) Header(id oid.ObjectID) (typ.Type, int, error) {
	ty, size, err := fetcher.store.ReadHeader(id)
	if err != nil {
		return typ.Unknown, 0, wrapObjectReadError(id, err)
	}

	return ty, size, nil
}

// Size returns the object content size at id.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) Size(id oid.ObjectID) (int, error) {
	size, err := fetcher.store.ReadSize(id)
	if err != nil {
		return 0, wrapObjectReadError(id, err)
	}

	return size, nil
}
