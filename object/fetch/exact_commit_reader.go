package fetch

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ExactCommitReader returns a reader for the content of the commit at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
//
// Labels: Life-Parent, Close-Caller.
func (r *Fetcher) ExactCommitReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeCommit, "commit")
}
