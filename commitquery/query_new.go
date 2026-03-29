package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectfetch "codeberg.org/lindenii/furgit/object/fetch"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// newQuery builds one empty mutable worker over one object fetcher and graph.
//
// Labels: Deps-Borrowed, Life-Parent.
func newQuery(fetcher *objectfetch.Fetcher, graph *commitgraphread.Reader) *query {
	return &query{
		fetcher:    fetcher,
		graph:      graph,
		byOID:      make(map[objectid.ObjectID]nodeIndex),
		byGraphPos: make(map[commitgraphread.Position]nodeIndex),
	}
}
