package commitquery

import (
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	objectid "lindenii.org/go/furgit/object/id"
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
