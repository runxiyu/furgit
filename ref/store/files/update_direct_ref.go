package files

import objectid "lindenii.org/go/furgit/object/id"

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
