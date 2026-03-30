package tree

var fileModeTable = map[FileMode]fileModeDetails{ //nolint:gochecknoglobals
	FileModeDir: {
		isBlobLike:    false,
		isRegularFile: false,
	},
	FileModeRegular: {
		isBlobLike:    true,
		isRegularFile: true,
	},
	FileModeExecutable: {
		isBlobLike:    true,
		isRegularFile: true,
	},
	FileModeSymlink: {
		isBlobLike:    true,
		isRegularFile: false,
	},
	FileModeGitlink: {
		isBlobLike:    false,
		isRegularFile: false,
	},
}
