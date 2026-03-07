package files

import "codeberg.org/lindenii/furgit/objectid"

type directKind uint8

const (
	directMissing directKind = iota
	directDetached
	directSymbolic
)

type directRef struct {
	kind     directKind
	name     string
	id       objectid.ObjectID
	target   string
	isLoose  bool
	isPacked bool
}
