package blob

import (
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/object/typ"
)

// BytesWithoutHeader renders the raw blob body bytes.
func (blob *Blob) AppendWithoutHeader(dst []byte) ([]byte, error) {
	return append(dst, blob.Data...), nil
}

// BytesWithHeader renders the raw object (header + body).
func (blob *Blob) AppendWithHeader(dst []byte) ([]byte, error) {
	dst = header.Append(dst, typ.TypeBlob, uint64(len(blob.Data)))

	return blob.AppendWithoutHeader(dst)
}
