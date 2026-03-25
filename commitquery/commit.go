package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// commitData stores the metadata needed by commit-domain queries.
type commitData struct {
	ID            objectid.ObjectID
	Parents       []parentRef
	CommitTime    int64
	Generation    uint64
	HasGeneration bool
	GraphPos      commitgraphread.Position
	HasGraphPos   bool
}
