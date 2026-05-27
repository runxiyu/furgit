package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// id returns one node's object ID.
func (query *query) id(idx nodeIndex) objectid.ObjectID {
	return query.nodes[idx].id
}
