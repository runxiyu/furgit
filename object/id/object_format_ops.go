package id

import "hash"

// HexLen returns the encoded hexadecimal length.
func (objectFormat ObjectFormat) HexLen() int {
	return objectFormat.Size() * 2
}

// Size returns the hash size in bytes.
func (objectFormat ObjectFormat) Size() int {
	return objectFormat.details().size
}

// New returns a new hash.Hash for this object format.
func (objectFormat ObjectFormat) New() (hash.Hash, error) {
	newFn := objectFormat.details().new
	if newFn == nil {
		return nil, ErrInvalidObjectFormat
	}

	return newFn(), nil
}

// String returns the canonical object format name.
func (objectFormat ObjectFormat) String() string {
	return objectFormat.details().name
}

// Sum computes an object ID from raw data using the selected object format.
func (objectFormat ObjectFormat) Sum(data []byte) ObjectID {
	return objectFormat.details().sum(data)
}

// Zero returns the all-zero object ID for this object format.
func (objectFormat ObjectFormat) Zero() ObjectID {
	return ObjectID{
		objectFormat: objectFormat,
		data:         [maxObjectIDSize]byte{},
	}
}
