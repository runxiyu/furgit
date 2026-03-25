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
