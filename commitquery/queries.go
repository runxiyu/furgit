package commitquery

import (
	"sync"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectfetch "codeberg.org/lindenii/furgit/object/fetch"
)

// Queries provides commit-domain queries over one object fetcher
// and optional commit-graph reader.
//
// Queries reuses internal mutable query workers across operations.
//
// Labels: MT-Safe.
type Queries struct {
	fetcher *objectfetch.Fetcher
	graph   *commitgraphread.Reader

	mu      sync.Mutex
	idle    []*query
	maxIdle int
}

// TODO: Research a shared arena, or perhaps worker-reconciliation
// schemes if a complete shared arena proves to be too contentious.
