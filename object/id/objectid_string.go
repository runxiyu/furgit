package objectid

import "encoding/hex"

// String returns the canonical hex representation.
func (id ObjectID) String() string {
	size := id.Algorithm().Size()

	return hex.EncodeToString(id.data[:size])
}
