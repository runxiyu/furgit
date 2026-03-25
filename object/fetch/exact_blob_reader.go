package fetch

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ExactBlobReader returns a reader for the content of the blob at id,
// together with its content size in bytes.
func (r *Fetcher) ExactBlobReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeBlob, "blob")
}
