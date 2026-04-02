package id

// ObjectID represents a Git object ID.
//
//nolint:recvcheck
type ObjectID struct {
	algo Algorithm
	data [maxObjectIDSize]byte
}
