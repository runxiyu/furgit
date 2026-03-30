package tree

// HasSameType reports whether mode and other describe the same tree entry kind.
//
// Regular files and executable files have the same type for diff-status purposes.
func (mode FileMode) HasSameType(other FileMode) bool {
	if mode == other {
		return true
	}

	return mode.details().isRegularFile && other.details().isRegularFile
}
