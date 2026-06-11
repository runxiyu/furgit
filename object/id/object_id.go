package id

// ObjectID represents a Git object ID.
//
//nolint:recvcheck
type ObjectID struct {
	objectFormat ObjectFormat
	data         [MaxObjectIDSize]byte
}
