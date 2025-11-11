package furgit

// Blob represents the contents of a Git blob.
type Blob struct {
	objectBase

	Data []byte
}

// ObjType allows Blob to satisfy the Object interface.
func (*Blob) ObjType() ObjType {
	return ObjBlob
}

func parseBlob(id Hash, body []byte) (*Blob, error) {
	data := append([]byte(nil), body...)
	return &Blob{
		objectBase: objectBase{Hash: id},
		Data:       data,
	}, nil
}

// Serialize renders the full "blob size\\0body" representation.
func (b *Blob) Serialize() ([]byte, error) {
	header, err := headerForType(ObjBlob, b.Data)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(b.Data))
	copy(raw, header)
	copy(raw[len(header):], b.Data)
	return raw, nil
}
