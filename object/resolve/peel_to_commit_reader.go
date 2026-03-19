package resolve

import (
	"io"

	"codeberg.org/lindenii/furgit/objectid"
)

// PeelToCommitReader returns a reader for the content of the peeled commit at
// id, together with its content size in bytes.
//
// Usage of this method is unusual.
func (r *Resolver) PeelToCommitReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	commitID, err := r.PeelToCommitID(id)
	if err != nil {
		return nil, 0, err
	}

	return r.ExactCommitReader(commitID)
}
