package files

import "codeberg.org/lindenii/furgit/objectid"

type directRefKind uint8

const (
	directMissing directRefKind = iota
	directDetached
	directSymbolic
)

type directRefState struct {
	kind     directRefKind
	name     string
	id       objectid.ObjectID
	target   string
	isLoose  bool
	isPacked bool
}
