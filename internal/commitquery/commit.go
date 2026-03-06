package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
)

// Commit stores the metadata needed by commit-domain queries.
type Commit struct {
	ID            objectid.ObjectID
	Parents       []Parent
	CommitTime    int64
	Generation    uint64
	HasGeneration bool
	GraphPos      commitgraphread.Position
	HasGraphPos   bool
}
