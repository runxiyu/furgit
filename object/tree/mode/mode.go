package mode

// Mode represents the mode of an entry in a Git tree.
type Mode uint32

const (
	// Directory is the mode of a subtree entry.
	Directory Mode = 0o40000

	// Regular is the mode of a non-executable regular file.
	Regular Mode = 0o100644

	// Executable is the mode of an executable regular file.
	Executable Mode = 0o100755

	// Symlink is the mode of a symbolic link.
	Symlink Mode = 0o120000

	// Gitlink is the mode of a gitlink (submodule commit) entry.
	Gitlink Mode = 0o160000
)

// maxModeDigits is the largest number of octal digits in any canonical mode.
const maxModeDigits = 6
