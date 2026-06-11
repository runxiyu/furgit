package blob

import (
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/typ"
)

// AppendWithoutHeader renders the raw blob body bytes.
func (blob *Blob) AppendWithoutHeader(dst []byte) ([]byte, error) {
	return append(dst, blob.Data...), nil
}

// AppendWithHeader renders the raw object (header + body).
func (blob *Blob) AppendWithHeader(dst []byte) ([]byte, error) {
	dst = header.Append(dst, typ.Blob, uint64(len(blob.Data)))

	return blob.AppendWithoutHeader(dst)
}
