package commitquery

import objectid "codeberg.org/lindenii/furgit/object/id"

// resolveOID resolves one commit object ID to one internal query node.
func (query *query) resolveOID(id objectid.ObjectID) (nodeIndex, error) {
	idx, ok := query.byOID[id]
	if ok {
		err := query.ensureLoaded(idx)
		if err != nil {
			return 0, err
		}

		return idx, nil
	}

	idx = query.newNode(id)
	query.byOID[id] = idx

	err := query.loadByOID(idx)
	if err != nil {
		delete(query.byOID, id)

		return 0, err
	}

	return idx, nil
}
