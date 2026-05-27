package files

import objectid "lindenii.org/go/furgit/object/id"

type queuedUpdate struct {
	name      string
	kind      updateKind
	newID     objectid.ObjectID
	oldID     objectid.ObjectID
	newTarget string
	oldTarget string
}
