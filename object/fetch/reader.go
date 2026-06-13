package fetch

import (
	"io"

	"lindenii.org/go/furgit/errs"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// exactReader reads one object's content stream
// and verifies that its header type matches wantType.
func (fetcher *Fetcher) exactReader(id oid.ObjectID, wantType typ.Type) (io.ReadCloser, int, error) {
	gotType, size, rc, err := fetcher.store.ReadReaderContent(id)
	if err != nil {
		return nil, 0, wrapObjectReadError(id, err)
	}

	if gotType != wantType {
		_ = rc.Close()

		return nil, 0, &errs.ObjectTypeError{OID: id, Got: gotType, Want: wantType}
	}

	return rc, size, nil
}
