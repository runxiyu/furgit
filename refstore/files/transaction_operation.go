package files

import "codeberg.org/lindenii/furgit/objectid"

type txOp struct {
	name      string
	kind      txKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}

type preparedTxOp struct {
	op     txOp
	target resolvedWriteTarget
}

type resolvedWriteTarget struct {
	name string
	loc  refPath
	ref  directRef
}
