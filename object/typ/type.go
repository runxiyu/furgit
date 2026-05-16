package typ

// Type represents a Git object type.
type Type uint8

const (
	// TypeBlob represents a Git blob.
	TypeBlob Type = iota

	// TypeTree represents a Git tree.
	TypeTree

	// TypeCommit represents a Git commit.
	TypeCommit

	// TypeTag represents a Git tag.
	TypeTag
)
