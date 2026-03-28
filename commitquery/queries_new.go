package commitquery

import (
	"runtime"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// New builds one concurrent-safe commit query service over one object store
// and optional commit-graph reader.
//
// Labels: Deps-Borrowed.
func New(store objectstore.ReadingStore, graph *commitgraphread.Reader) *Queries {
	maxIdle := runtime.GOMAXPROCS(0)
	if maxIdle < 1 {
		maxIdle = 1
	}

	return &Queries{
		store:   store,
		graph:   graph,
		maxIdle: maxIdle,
	}
}
