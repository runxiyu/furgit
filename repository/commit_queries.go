package repository

import "codeberg.org/lindenii/furgit/commitquery"

// CommitQueries returns commit queries backed by the repository's object store
// and optional commit-graph.
//
// Use CommitQueries for ancestor checks and merge-base computation.
//
// Labels: Life-Parent.
func (repo *Repository) CommitQueries() *commitquery.Queries {
	return repo.commitQueries
}
