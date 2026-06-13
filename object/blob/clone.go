package blob

import "bytes"

// Clone returns a deep copy of the blob
// whose Data is independent of any memory the original may alias.
//
// Labels: Life-Independent.
func (blob *Blob) Clone() *Blob {
	return &Blob{Data: bytes.Clone(blob.Data)}
}
