package commitquery

import commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"

// ResolveGraphPos resolves one commit-graph position to one internal query node.
func (ctx *Context) ResolveGraphPos(pos commitgraphread.Position) (NodeIndex, error) {
	idx, ok := ctx.byGraphPos[pos]
	if ok {
		err := ctx.ensureLoaded(idx)
		if err != nil {
			return 0, err
		}

		return idx, nil
	}

	commit, err := ctx.graph.CommitAt(pos)
	if err != nil {
		return 0, err
	}

	idx, ok = ctx.byOID[commit.OID]
	if !ok {
		idx = ctx.newNode(commit.OID)
		ctx.byOID[commit.OID] = idx
	}

	ctx.byGraphPos[pos] = idx
	ctx.nodes[idx].graphPos = pos
	ctx.nodes[idx].hasGraphPos = true

	err = ctx.loadCommitAtGraphPos(idx, pos)
	if err != nil {
		delete(ctx.byGraphPos, pos)

		return 0, err
	}

	return idx, nil
}

// loadByGraphPos populates one node from a commit-graph position.
func (ctx *Context) loadByGraphPos(idx NodeIndex) error {
	pos := ctx.nodes[idx].graphPos

	return ctx.loadCommitAtGraphPos(idx, pos)
}

func (ctx *Context) loadCommitAtGraphPos(idx NodeIndex, pos commitgraphread.Position) error {
	commit, err := ctx.graph.CommitAt(pos)
	if err != nil {
		return err
	}

	parents := make([]Parent, 0, 2+len(commit.ExtraParents))

	if commit.Parent1.Valid {
		parentOID, err := ctx.graph.OIDAt(commit.Parent1.Pos)
		if err != nil {
			return err
		}

		parents = append(parents, Parent{
			ID:          parentOID,
			GraphPos:    commit.Parent1.Pos,
			HasGraphPos: true,
		})
	}

	if commit.Parent2.Valid {
		parentOID, err := ctx.graph.OIDAt(commit.Parent2.Pos)
		if err != nil {
			return err
		}

		parents = append(parents, Parent{
			ID:          parentOID,
			GraphPos:    commit.Parent2.Pos,
			HasGraphPos: true,
		})
	}

	for _, parentPos := range commit.ExtraParents {
		parentOID, err := ctx.graph.OIDAt(parentPos)
		if err != nil {
			return err
		}

		parents = append(parents, Parent{
			ID:          parentOID,
			GraphPos:    parentPos,
			HasGraphPos: true,
		})
	}

	data := Commit{
		ID:            commit.OID,
		Parents:       parents,
		CommitTime:    commit.CommitTimeUnix,
		Generation:    commit.GenerationV2,
		HasGeneration: commit.GenerationV2 != 0,
		GraphPos:      pos,
		HasGraphPos:   true,
	}

	return ctx.populateNode(idx, data)
}
