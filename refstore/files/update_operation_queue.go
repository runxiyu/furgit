package files

import "codeberg.org/lindenii/furgit/objectid"

type queuedUpdate struct {
	name      string
	kind      updateKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}
