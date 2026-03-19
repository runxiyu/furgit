package resolve

import (
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ExactCommitReader returns a reader for the content of the commit at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
func (r *Resolver) ExactCommitReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeCommit, "commit")
}
