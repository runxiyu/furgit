package blob

import (
	"errors"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// BytesWithoutHeader renders the raw blob body bytes.
func (blob *Blob) BytesWithoutHeader() ([]byte, error) {
	return append([]byte(nil), blob.Data...), nil
}

// BytesWithHeader renders the raw object (header + body).
func (blob *Blob) BytesWithHeader() ([]byte, error) {
	body, err := blob.BytesWithoutHeader()
	if err != nil {
		return nil, err
	}

	header, ok := objectheader.Encode(objecttype.TypeBlob, int64(len(body)))
	if !ok {
		return nil, errors.New("object: blob: failed to encode object header")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw, nil
}
