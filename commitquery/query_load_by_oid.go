package commitquery

import (
	stderrors "errors"

	giterrors "codeberg.org/lindenii/furgit/errors"
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/object/commit"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// loadByOID populates one node from an object ID.
func (query *query) loadByOID(idx nodeIndex) error {
	id := query.nodes[idx].id

	if query.graph != nil {
		pos, err := query.graph.Lookup(id)
		if err != nil {
			if _, ok := stderrors.AsType[*commitgraphread.NotFoundError](err); !ok {
				return err
			}
		} else {
			return query.loadCommitAtGraphPos(idx, pos)
		}
	}

	obj, err := query.fetcher.ExactObject(id)
	if err != nil {
		if stderrors.Is(err, objectstore.ErrObjectNotFound) {
			return &giterrors.ObjectMissingError{OID: id}
		}

		return err
	}

	commitObj, ok := obj.Object().(*commit.Commit)
	if !ok {
		return &giterrors.ObjectTypeError{
			OID:  id,
			Got:  obj.Object().ObjectType(),
			Want: objecttype.TypeCommit,
		}
	}

	parents := make([]parentRef, 0, len(commitObj.Parents))
	for _, parentID := range commitObj.Parents {
		parents = append(parents, parentRef{ID: parentID})
	}

	commit := commitData{
		ID:         id,
		Parents:    parents,
		CommitTime: commitObj.Committer.WhenUnix,
	}

	return query.populateNode(idx, commit)
}
