package resolve

import (
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ExactBlobReader returns a reader for the content of the blob at id,
// together with its content size in bytes.
func (r *Resolver) ExactBlobReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeBlob, "blob")
}
