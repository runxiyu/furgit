package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// newNode allocates one empty internal node.
func (query *query) newNode(id objectid.ObjectID) nodeIndex {
	count := len(query.nodes)

	idx := nodeIndex(count)

	query.nodes = append(query.nodes, node{id: id})

	return idx
}
