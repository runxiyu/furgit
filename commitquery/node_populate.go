package commitquery

import "fmt"

// populateNode fills one node's metadata and resolves its parents.
func (query *query) populateNode(idx nodeIndex, commit commitData) error {
	if query.nodes[idx].loaded {
		if query.nodes[idx].id != commit.ID {
			return fmt.Errorf("commitquery: node identity mismatch: have %s, got %s", query.nodes[idx].id, commit.ID)
		}

		return nil
	}

	query.nodes[idx].id = commit.ID
	query.nodes[idx].commitTime = commit.CommitTime
	query.nodes[idx].generation = commit.Generation
	query.nodes[idx].hasGeneration = commit.HasGeneration

	if commit.HasGraphPos {
		query.nodes[idx].graphPos = commit.GraphPos
		query.nodes[idx].hasGraphPos = true
		query.byGraphPos[commit.GraphPos] = idx
	}

	query.nodes[idx].loaded = true
	query.nodes[idx].parents = query.nodes[idx].parents[:0]

	for _, parent := range commit.Parents {
		parentIdx, err := query.resolveParent(parent)
		if err != nil {
			query.nodes[idx].loaded = false
			query.nodes[idx].parents = nil

			return err
		}

		query.nodes[idx].parents = append(query.nodes[idx].parents, parentIdx)
	}

	return nil
}
