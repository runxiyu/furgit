package testgit

import (
	"testing"

	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
)

// OpenCommitGraph opens the repository commit-graph and registers cleanup on
// the caller.
func (testRepo *TestRepo) OpenCommitGraph(tb testing.TB) *commitgraphread.Reader {
	tb.Helper()

	objectsRoot := testRepo.OpenObjectsRoot(tb)

	graph, err := commitgraphread.Open(objectsRoot, testRepo.Algorithm(), commitgraphread.OpenSingle)
	if err != nil {
		tb.Fatalf("commitgraphread.Open: %v", err)
	}

	tb.Cleanup(func() {
		_ = graph.Close()
	})

	return graph
}
