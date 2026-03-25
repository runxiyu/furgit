package resolve

import (
	"fmt"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// exactReader reads one object's content stream and verifies that its header
// type matches wantType.
func (r *Resolver) exactReader(id objectid.ObjectID, wantType objecttype.Type, wantName string) (io.ReadCloser, int64, error) {
	gotType, size, rc, err := r.store.ReadReaderContent(id)
	if err != nil {
		return nil, 0, err
	}

	if gotType != wantType {
		_ = rc.Close()

		return nil, 0, fmt.Errorf("object/resolve: expected %s object %s, got %v", wantName, id, gotType)
	}

	return rc, size, nil
}
