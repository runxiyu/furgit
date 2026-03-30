package tree

type fileModeDetails struct {
	isBlobLike    bool
	isRegularFile bool
}

func (mode FileMode) details() fileModeDetails {
	return fileModeTable[mode]
}
