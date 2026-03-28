package commitquery

import commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"

// resolveGraphPos resolves one commit-graph position to one internal query node.
func (query *query) resolveGraphPos(pos commitgraphread.Position) (nodeIndex, error) {
	idx, ok := query.byGraphPos[pos]
	if ok {
		err := query.ensureLoaded(idx)
		if err != nil {
			return 0, err
		}

		return idx, nil
	}

	commit, err := query.graph.CommitAt(pos)
	if err != nil {
		return 0, err
	}

	idx, ok = query.byOID[commit.OID]
	if !ok {
		idx = query.newNode(commit.OID)
		query.byOID[commit.OID] = idx
	}

	query.byGraphPos[pos] = idx
	query.nodes[idx].graphPos = pos
	query.nodes[idx].hasGraphPos = true

	err = query.loadCommitAtGraphPos(idx, pos)
	if err != nil {
		delete(query.byGraphPos, pos)

		return 0, err
	}

	return idx, nil
}

// loadByGraphPos populates one node from a commit-graph position.
func (query *query) loadByGraphPos(idx nodeIndex) error {
	pos := query.nodes[idx].graphPos

	return query.loadCommitAtGraphPos(idx, pos)
}

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
