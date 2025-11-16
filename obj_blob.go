package furgit

// Blob represents the contents of a Git blob.
type Blob[T HashType] struct {
	Hash Hash[T]
	Data []byte
}

// ObjType allows Blob to satisfy the Object interface.
func (*Blob[T]) ObjType() ObjType {
	return ObjBlob
}

func parseBlob[T HashType](id Hash[T], body []byte) (*Blob[T], error) {
	data := append([]byte(nil), body...)
	return &Blob[T]{
		Hash: id,
		Data: data,
	}, nil
}

// Serialize renders the full "blob size\\0body" representation.
func (b *Blob[T]) Serialize() ([]byte, error) {
	header, err := headerForType(ObjBlob, b.Data)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(b.Data))
	copy(raw, header)
	copy(raw[len(header):], b.Data)
	return raw, nil
}
