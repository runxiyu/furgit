package typ

// Type represents a Git object type.
type Type uint8

const (
	// TypeInvalid represents an invalid Git object type.
	TypeInvalid Type = 0

	// TypeCommit represents a Git commit.
	TypeCommit Type = 1

	// TypeTree represents a Git tree.
	TypeTree Type = 2

	// TypeBlob represents a Git blob.
	TypeBlob Type = 3

	// TypeTag represents a Git tag.
	TypeTag Type = 4
)
