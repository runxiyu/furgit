package repository

import "codeberg.org/lindenii/furgit/reachability"

// Reachability returns graph traversal helpers backed by the repository's
// object store and optional commit-graph.
//
// Use Reachability to walk reachable commits or objects and to perform
// connectivity checks.
//
// Labels: Life-Parent.
func (repo *Repository) Reachability() *reachability.Reachability {
	return reachability.New(repo.fetcher, repo.commitGraph)
}
