package fetch

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// PeelToCommitReader returns a reader for the content of the peeled commit at
// id, together with its content size in bytes.
//
// Usage of this method is unusual.
//
// Labels: Life-Parent, Close-Caller.
func (r *Fetcher) PeelToCommitReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	commitID, err := r.PeelToCommitID(id)
	if err != nil {
		return nil, 0, err
	}

	return r.ExactCommitReader(commitID)
}
