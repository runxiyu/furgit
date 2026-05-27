package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// compare orders two internal nodes using merge-base queue ordering.
func (query *query) compare(left, right nodeIndex) int {
	leftGeneration := query.effectiveGeneration(left)
	rightGeneration := query.effectiveGeneration(right)

	switch {
	case leftGeneration < rightGeneration:
		return -1
	case leftGeneration > rightGeneration:
		return 1
	}

	switch {
	case query.nodes[left].commitTime < query.nodes[right].commitTime:
		return -1
	case query.nodes[left].commitTime > query.nodes[right].commitTime:
		return 1
	}

	return objectid.Compare(query.nodes[left].id, query.nodes[right].id)
}
