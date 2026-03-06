package reachability

import (
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}
