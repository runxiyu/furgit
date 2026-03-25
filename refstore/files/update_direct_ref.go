package files

import objectid "codeberg.org/lindenii/furgit/object/id"

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
