package object

import "codeberg.org/lindenii/furgit/objecttype"

// Serialize renders the raw object (header + body).
func (blob *Blob) Serialize() ([]byte, error) {
	header, err := headerForType(objecttype.TypeBlob, blob.Data)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(blob.Data))
	copy(raw, header)
	copy(raw[len(header):], blob.Data)
	return raw, nil
}
