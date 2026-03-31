package commitquery

import (
	stderrors "errors"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
)

// loadByOID populates one node from an object ID.
func (query *query) loadByOID(idx nodeIndex) error {
	id := query.nodes[idx].id

	if query.graph != nil {
		pos, err := query.graph.Lookup(id)
		if err != nil {
			if _, ok := stderrors.AsType[*commitgraphread.NotFoundError](err); !ok {
				return err
			}
		} else {
			return query.loadCommitAtGraphPos(idx, pos)
		}
	}

	commit, err := query.fetcher.ExactCommit(id)
	if err != nil {
		return err
	}

	parents := make([]parentRef, 0, len(commit.Object().Parents))
	for _, parentID := range commit.Object().Parents {
		parents = append(parents, parentRef{ID: parentID})
	}

	commitData := commitData{
		ID:         id,
		Parents:    parents,
		CommitTime: commit.Object().Committer.WhenUnix,
	}

	return query.populateNode(idx, commitData)
}
