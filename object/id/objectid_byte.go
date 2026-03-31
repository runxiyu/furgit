package objectid

// Bytes returns a copy of the object ID bytes.
func (id ObjectID) Bytes() []byte {
	size := id.Algorithm().Size()

	return append([]byte(nil), id.data[:size]...)
}

// RawBytes returns a direct byte slice view of the object ID bytes.
//
// Use Bytes when an independent copy is required.
//
// Labels: Mut-Never.
func (id *ObjectID) RawBytes() []byte {
	size := id.Algorithm().Size()

	return id.data[:size:size]
}
