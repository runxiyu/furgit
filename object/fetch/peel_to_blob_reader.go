package fetch

import (
	"io"

	objectid "lindenii.org/go/furgit/object/id"
)

// PeelToBlobReader returns a reader for the content of the peeled blob at id,
// together with its content size in bytes.
//
// Labels: Life-Parent, Close-Caller.
func (r *Fetcher) PeelToBlobReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	blobID, err := r.PeelToBlobID(id)
	if err != nil {
		return nil, 0, err
	}

	return r.ExactBlobReader(blobID)
}
