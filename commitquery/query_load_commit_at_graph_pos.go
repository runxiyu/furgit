package commitquery

import commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"

// loadCommitAtGraphPos populates one node from one commit-graph record.
func (query *query) loadCommitAtGraphPos(idx nodeIndex, pos commitgraphread.Position) error {
	commit, err := query.graph.CommitAt(pos)
	if err != nil {
		return err
	}

	parents := make([]parentRef, 0, 2+len(commit.ExtraParents))

	if commit.Parent1.Valid {
		parentOID, err := query.graph.OIDAt(commit.Parent1.Pos)
		if err != nil {
			return err
		}

		parents = append(parents, parentRef{
			ID:          parentOID,
			GraphPos:    commit.Parent1.Pos,
			HasGraphPos: true,
		})
	}

	if commit.Parent2.Valid {
		parentOID, err := query.graph.OIDAt(commit.Parent2.Pos)
		if err != nil {
			return err
		}

		parents = append(parents, parentRef{
			ID:          parentOID,
			GraphPos:    commit.Parent2.Pos,
			HasGraphPos: true,
		})
	}

	for _, parentPos := range commit.ExtraParents {
		parentOID, err := query.graph.OIDAt(parentPos)
		if err != nil {
			return err
		}

		parents = append(parents, parentRef{
			ID:          parentOID,
			GraphPos:    parentPos,
			HasGraphPos: true,
		})
	}

	data := commitData{
		ID:            commit.OID,
		Parents:       parents,
		CommitTime:    commit.CommitTimeUnix,
		Generation:    commit.GenerationV2,
		HasGeneration: commit.GenerationV2 != 0,
		GraphPos:      pos,
		HasGraphPos:   true,
	}

	return query.populateNode(idx, data)
}
