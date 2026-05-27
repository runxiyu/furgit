package commitquery

import (
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectid "lindenii.org/go/furgit/object/id"
)

// parentRef references one commit parent.
type parentRef struct {
	ID          objectid.ObjectID
	GraphPos    commitgraphread.Position
	HasGraphPos bool
}
