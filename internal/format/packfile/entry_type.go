package packfile

import (
	"errors"

	"lindenii.org/go/furgit/object/typ"
)

var (
	// ErrInternalEntryType reports that
	// a supplied packfile [EntryType]
	// cannot be converted into an ordinary [typ.Type]
	// since it is a packfile implementation detail.
	ErrInternalEntryType = errors.New("internal/format/packfile: packfile-internal entry type cannot be converted to an object type")

	// ErrUnrepresentableObjectType reports that
	// a supplied ordinary [typ.Type]
	// is not currently representable
	// as a packfile [EntryType].
	ErrUnrepresentableObjectType = errors.New("internal/format/packfile: object type not representable in packfiles")
)

// EntryType represents the type of an entry in a git packfile.
type EntryType uint8

const (
	EntryTypeInvalid  EntryType = 0
	EntryTypeCommit   EntryType = 1
	EntryTypeTree     EntryType = 2
	EntryTypeBlob     EntryType = 3
	EntryTypeTag      EntryType = 4
	EntryTypeFuture   EntryType = 5
	EntryTypeOfsDelta EntryType = 6
	EntryTypeRefDelta EntryType = 7
)

// ObjectTypeToEntryType converts an ordinary [typ.Type] into a packfile [EntryType].
func ObjectTypeToEntryType(ty typ.Type) (EntryType, error) {
	switch ty {
	case typ.Commit:
		return EntryTypeCommit, nil
	case typ.Tree:
		return EntryTypeTree, nil
	case typ.Blob:
		return EntryTypeBlob, nil
	case typ.Tag:
		return EntryTypeTag, nil
	}

	return EntryTypeInvalid, ErrUnrepresentableObjectType
}

// ObjectTypeToEntryType converts a a packfile [EntryType] into an ordinary [typ.Type].
func (entryType EntryType) EntryTypeToObjectType() (typ.Type, error) {
	switch entryType {
	case EntryTypeCommit:
		return typ.Commit, nil
	case EntryTypeTree:
		return typ.Tree, nil
	case EntryTypeBlob:
		return typ.Blob, nil
	case EntryTypeTag:
		return typ.Tag, nil
	}

	return typ.Unknown, ErrInternalEntryType
}
