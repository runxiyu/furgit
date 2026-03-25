package resolve

import (
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ExactTreeReader returns a reader for the content of the tree at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
func (r *Resolver) ExactTreeReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	return r.exactReader(id, objecttype.TypeTree, "tree")
}
