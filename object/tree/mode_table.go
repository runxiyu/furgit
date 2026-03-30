package tree

var fileModeTable = map[FileMode]fileModeDetails{ //nolint:gochecknoglobals
	FileModeDir: {
		isBlobLike: false,
	},
	FileModeRegular: {
		isBlobLike: true,
	},
	FileModeExecutable: {
		isBlobLike: true,
	},
	FileModeSymlink: {
		isBlobLike: true,
	},
	FileModeGitlink: {
		isBlobLike: false,
	},
}
