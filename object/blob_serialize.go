package object

import (
	"errors"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// SerializeWithoutHeader renders the raw blob body bytes.
func (blob *Blob) SerializeWithoutHeader() ([]byte, error) {
	return append([]byte(nil), blob.Data...), nil
}

// SerializeWithHeader renders the raw object (header + body).
func (blob *Blob) SerializeWithHeader() ([]byte, error) {
	body, err := blob.SerializeWithoutHeader()
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
