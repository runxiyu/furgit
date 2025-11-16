package furgit

// Blob represents the contents of a Git blob.
type Blob struct {
	Data []byte
}

// StoredBlob represents a blob stored in the object database.
type StoredBlob struct {
	Blob
	hash Hash
}

// Hash returns the hash of the stored blob.
func (sBlob *StoredBlob) Hash() Hash {
	return sBlob.hash
}

// ObjectType allows Blob to satisfy the Object interface.
func (blob *Blob) ObjectType() ObjectType {
	_ = blob
	return ObjectTypeBlob
}

func parseBlob(id Hash, body []byte) (*StoredBlob, error) {
	data := append([]byte(nil), body...)
	return &StoredBlob{
		hash: id,
		Blob: Blob{
			Data: data,
		},
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
