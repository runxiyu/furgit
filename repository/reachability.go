package repository

import "codeberg.org/lindenii/furgit/reachability"

// Reachability returns graph traversal helpers backed by the repository's
// object store and optional commit-graph.
//
// Labels: Life-Parent, Close-No.
func (repo *Repository) Reachability() *reachability.Reachability {
	return reachability.New(repo.objects, repo.commitGraph)
}
