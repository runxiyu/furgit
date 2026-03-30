package tree

type fileModeDetails struct {
	isBlobLike bool
}

func (mode FileMode) details() fileModeDetails {
	return fileModeTable[mode]
}
