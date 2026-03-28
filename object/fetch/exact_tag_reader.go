package fetch

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ExactTagReader returns a reader for the content of the tag at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
//
// Labels: Life-Parent, Close-Caller.
func (r *Fetcher) ExactTagReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeTag, "tag")
}
