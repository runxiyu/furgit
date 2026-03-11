package commitquery

import (
	"math"

	"codeberg.org/lindenii/furgit/objectid"
)

// EffectiveGeneration returns one node's generation value.
func (query *Query) effectiveGeneration(idx nodeIndex) uint64 {
	if !query.nodes[idx].hasGeneration {
		return generationInfinity
	}

	return query.nodes[idx].generation
}

const (
	generationInfinity = uint64(math.MaxUint64)
)

func compareByGeneration(query *Query) func(nodeIndex, nodeIndex) int {
	return func(left, right nodeIndex) int {
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
}
