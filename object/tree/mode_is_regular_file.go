package tree

// IsRegularFile reports whether mode names one regular-file variant.
func (mode FileMode) IsRegularFile() bool {
	return mode.details().isRegularFile
}
