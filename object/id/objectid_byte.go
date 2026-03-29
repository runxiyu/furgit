package objectid

// Bytes returns a copy of the object ID bytes.
func (id ObjectID) Bytes() []byte {
	size := id.Algorithm().Size()

	return append([]byte(nil), id.data[:size]...)
}

// RawBytes returns a direct byte slice view of the object ID bytes.
//
// The returned slice aliases the object ID's internal storage. Callers MUST
// treat it as read-only and MUST NOT modify its contents.
//
// Use Bytes when an independent copy is required.
func (id *ObjectID) RawBytes() []byte {
	size := id.Algorithm().Size()

	return id.data[:size:size]
}
