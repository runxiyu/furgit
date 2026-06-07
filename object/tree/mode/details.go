package mode

import "lindenii.org/go/furgit/object/typ"

type modeDetails struct {
	valid         bool
	isBlobLike    bool
	isRegularFile bool
	objectType    typ.Type
}

func (mode Mode) details() modeDetails {
	return modeTable[mode]
}

//nolint:gochecknoglobals
var modeTable = map[Mode]modeDetails{
	Directory: {
		valid:         true,
		isBlobLike:    false,
		isRegularFile: false,
		objectType:    typ.TypeTree,
	},
	Regular: {
		valid:         true,
		isBlobLike:    true,
		isRegularFile: true,
		objectType:    typ.TypeBlob,
	},
	Executable: {
		valid:         true,
		isBlobLike:    true,
		isRegularFile: true,
		objectType:    typ.TypeBlob,
	},
	Symlink: {
		valid:         true,
		isBlobLike:    true,
		isRegularFile: false,
		objectType:    typ.TypeBlob,
	},
	Gitlink: {
		valid:         true,
		isBlobLike:    false,
		isRegularFile: false,
		objectType:    typ.TypeCommit,
	},
}
