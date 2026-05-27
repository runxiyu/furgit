package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// resolveCommitish peels one commit-ish object ID and resolves the commit.
func (query *query) resolveCommitish(id objectid.ObjectID) (nodeIndex, error) {
	id, err := query.fetcher.PeelToCommitID(id)
	if err != nil {
		return 0, err
	}

	return query.resolveOID(id)
}
