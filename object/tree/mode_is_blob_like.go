package tree

// IsBlobLike reports whether mode names one blob-like tree entry kind.
//
// Blob-like entries store blob object IDs as their targets.
func (mode FileMode) IsBlobLike() bool {
	return mode.details().isBlobLike
}
