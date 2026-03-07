package files

type txKind uint8

const (
	txCreate txKind = iota
	txUpdate
	txDelete
	txVerify
	txCreateSymbolic
	txUpdateSymbolic
	txDeleteSymbolic
	txVerifySymbolic
)
