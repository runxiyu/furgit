package commitquery

import (
	"sync"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Queries provides commit-domain queries over one object store
// and optional commit-graph reader.
//
// Queries reuses internal mutable query workers across operations.
//
// Labels: MT-Safe.
type Queries struct {
	store objectstore.ReadingStore
	graph *commitgraphread.Reader

	mu      sync.Mutex
	idle    []*query
	maxIdle int
}
