package fetch

import (
	"io"

	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ExactBlobReader returns a reader for the content of the blob at id,
// together with its content size in bytes.
//
// Labels: Life-Parent, Close-Caller.
func (r *Fetcher) ExactBlobReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeBlob)
}
