package trees

// EntryKind identifies a tree-diff entry kind.
type EntryKind int

const (
	// EntryKindInvalid indicates an invalid diff entry kind.
	EntryKindInvalid EntryKind = iota
	// EntryKindDeleted indicates a deleted path.
	EntryKindDeleted
	// EntryKindAdded indicates an added path.
	EntryKindAdded
	// EntryKindModified indicates a modified path.
	EntryKindModified
)
