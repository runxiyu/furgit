package commitquery

import "fmt"

// populateNode fills one node's metadata and resolves its parents.
func (ctx *Context) populateNode(idx NodeIndex, commit Commit) error {
	if ctx.nodes[idx].loaded {
		if ctx.nodes[idx].id != commit.ID {
			return fmt.Errorf("commitquery: node identity mismatch: have %s, got %s", ctx.nodes[idx].id, commit.ID)
		}

		return nil
	}

	ctx.nodes[idx].id = commit.ID
	ctx.nodes[idx].commitTime = commit.CommitTime
	ctx.nodes[idx].generation = commit.Generation
	ctx.nodes[idx].hasGeneration = commit.HasGeneration

	if commit.HasGraphPos {
		ctx.nodes[idx].graphPos = commit.GraphPos
		ctx.nodes[idx].hasGraphPos = true
		ctx.byGraphPos[commit.GraphPos] = idx
	}

	ctx.nodes[idx].loaded = true
	ctx.nodes[idx].parents = ctx.nodes[idx].parents[:0]

	for _, parent := range commit.Parents {
		parentIdx, err := ctx.resolveParent(parent)
		if err != nil {
			ctx.nodes[idx].loaded = false
			ctx.nodes[idx].parents = nil

			return err
		}

		ctx.nodes[idx].parents = append(ctx.nodes[idx].parents, parentIdx)
	}

	return nil
}
