package furgit

// Blob represents the contents of a Git blob.
type Blob struct {
	Hash Hash

	Data []byte
}

// ObjectType allows Blob to satisfy the Object interface.
func (blob *Blob) ObjectType() ObjectType {
	_ = blob
	return ObjectTypeBlob
}

func parseBlob(id Hash, body []byte) (*Blob, error) {
	data := append([]byte(nil), body...)
	return &Blob{
		Hash: id,
		Data: data,
	}, nil
}

// Serialize renders the full "blob size\\0body" representation.
func (blob *Blob) Serialize() ([]byte, error) {
	header, err := headerForType(ObjectTypeBlob, blob.Data)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(blob.Data))
	copy(raw, header)
	copy(raw[len(header):], blob.Data)
	return raw, nil
}
