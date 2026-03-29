package fetch

import (
	"io"

	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
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
