package commitquery

import (
	"runtime"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectfetch "codeberg.org/lindenii/furgit/object/fetch"
)

// New builds one concurrent-safe commit query service over one object fetcher
// and optional commit-graph reader.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(fetcher *objectfetch.Fetcher, graph *commitgraphread.Reader) *Queries {
	maxIdle := max(runtime.GOMAXPROCS(0), 1)

	return &Queries{
		fetcher: fetcher,
		graph:   graph,
		maxIdle: maxIdle,
	}
}
