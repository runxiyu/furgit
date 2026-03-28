package repository

import "codeberg.org/lindenii/furgit/commitquery"

// CommitQueries returns commit queries backed by the repository's object store
// and optional commit-graph.
//
// Labels: Life-Parent, Close-No.
func (repo *Repository) CommitQueries() *commitquery.Queries {
	return repo.commitQueries
}
