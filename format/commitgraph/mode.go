package commitgraph

// OpenMode controls which commit-graph layout Open loads.
type OpenMode uint8

const (
	// OpenSingle opens one commit-graph file at info/commit-graph.
	OpenSingle OpenMode = iota
	// OpenChain opens chained commit-graphs from info/commit-graphs.
	OpenChain
)
