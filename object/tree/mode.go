package tree

// FileMode represents the mode of a file in a Git tree.
type FileMode uint32

const (
	FileModeDir        FileMode = 0o40000
	FileModeRegular    FileMode = 0o100644
	FileModeExecutable FileMode = 0o100755
	FileModeSymlink    FileMode = 0o120000
	FileModeGitlink    FileMode = 0o160000
)

// IsBlobLike reports whether mode names one blob-like tree entry kind.
//
// Blob-like entries store blob object IDs as their targets.
func (mode FileMode) IsBlobLike() bool {
	switch mode {
	case FileModeRegular, FileModeExecutable, FileModeSymlink:
		return true
	default:
		return false
	}
}
