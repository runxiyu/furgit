package fetch

import (
	"io"

	giterrors "lindenii.org/go/furgit/errors"
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// exactReader reads one object's content stream and verifies that its header
// type matches wantType.
func (r *Fetcher) exactReader(id objectid.ObjectID, wantType objecttype.Type) (io.ReadCloser, int64, error) {
	gotType, size, rc, err := r.store.ReadReaderContent(id)
	if err != nil {
		return nil, 0, wrapObjectReadError(id, err)
	}

	if gotType != wantType {
		_ = rc.Close()

		return nil, 0, &giterrors.ObjectTypeError{OID: id, Got: gotType, Want: wantType}
	}

	return rc, size, nil
}
