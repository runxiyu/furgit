package reachability

import (
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}
