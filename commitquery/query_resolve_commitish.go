package commitquery

import (
	"codeberg.org/lindenii/furgit/internal/peel"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// resolveCommitish peels one commit-ish object ID and resolves the commit.
func (query *query) resolveCommitish(id objectid.ObjectID) (nodeIndex, error) {
	commitID, err := peel.ToCommit(query.store, id)
	if err != nil {
		return 0, err
	}

	return query.resolveOID(commitID)
}
