package resolve

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// PeelToTreeReader returns a reader for the content of the peeled tree at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
func (r *Resolver) PeelToTreeReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	treeID, err := r.PeelToTreeID(id)
	if err != nil {
		return nil, 0, err
	}

	return r.ExactTreeReader(treeID)
}
