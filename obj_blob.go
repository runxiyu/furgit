package furgit

// Blob represents a Git blob object.
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

// ObjectType returns the object type of the blob.
//
// It always returns ObjectTypeBlob.
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

// Serialize renders the blob into its raw byte representation,
// including the header (i.e., "type size\0").
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
