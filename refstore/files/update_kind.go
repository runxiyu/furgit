package files

type updateKind uint8

const (
	updateCreate updateKind = iota
	updateReplace
	updateDelete
	updateVerify
	updateCreateSymbolic
	updateReplaceSymbolic
	updateDeleteSymbolic
	updateVerifySymbolic
)
