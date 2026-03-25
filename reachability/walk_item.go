package reachability

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}
