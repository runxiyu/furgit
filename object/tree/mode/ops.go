package mode

import "lindenii.org/go/furgit/object/typ"

// IsValid reports whether mode is a canonical mode that Git writes.
func (mode Mode) IsValid() bool {
	return mode.details().valid
}

// IsBlobLike reports whether mode names one blob-like tree entry kind.
//
// Blob-like entries store blob object IDs as their targets.
func (mode Mode) IsBlobLike() bool {
	return mode.details().isBlobLike
}

// IsRegularFile reports whether mode names one regular-file variant.
func (mode Mode) IsRegularFile() bool {
	return mode.details().isRegularFile
}

// HasSameType reports whether mode and other describe the same tree entry kind.
//
// Regular files and executable files have the same type for diff-status purposes.
func (mode Mode) HasSameType(other Mode) bool {
	if mode == other {
		return true
	}

	return mode.details().isRegularFile && other.details().isRegularFile
}

// ObjectType returns the type of object that an entry with this mode targets.
//
// It returns [typ.Unknown] for invalid modes.
func (mode Mode) ObjectType() typ.Type {
	return mode.details().objectType
}
