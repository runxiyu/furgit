package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// parentRef references one commit parent.
type parentRef struct {
	ID          objectid.ObjectID
	GraphPos    commitgraphread.Position
	HasGraphPos bool
}
