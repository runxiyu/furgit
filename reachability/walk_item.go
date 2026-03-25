package reachability

import (
	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}
