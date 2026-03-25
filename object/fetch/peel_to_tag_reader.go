package fetch

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// PeelToTagReader returns a reader for the content of the tag at id,
// together with its content size in bytes.
//
// Usage of this method is unusual.
func (r *Fetcher) PeelToTagReader(id objectid.ObjectID) (io.ReadCloser, int64, error) {
	tagID, err := r.PeelToTagID(id)
	if err != nil {
		return nil, 0, err
	}

	return r.ExactTagReader(tagID)
}
