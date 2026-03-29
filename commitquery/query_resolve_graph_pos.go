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
