package files

import objectid "codeberg.org/lindenii/furgit/object/id"

type queuedUpdate struct {
	name      string
	kind      updateKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}
