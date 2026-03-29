package commitquery

import objectid "codeberg.org/lindenii/furgit/object/id"

func (query *query) id(idx nodeIndex) objectid.ObjectID {
	return query.nodes[idx].id
}
