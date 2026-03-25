package commitquery

import (
	stderrors "errors"

	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/internal/peel"
	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/objectstore"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (query *Query) id(idx nodeIndex) objectid.ObjectID {
	return query.nodes[idx].id
}

func (query *Query) commitTime(idx nodeIndex) int64 {
	return query.nodes[idx].commitTime
}

func (query *Query) resolveOID(id objectid.ObjectID) (nodeIndex, error) {
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

func (query *Query) resolveCommitish(id objectid.ObjectID) (nodeIndex, error) {
	commitID, err := peel.ToCommit(query.store, id)
	if err != nil {
		return 0, err
	}

	return query.resolveOID(commitID)
}

// loadByOID populates one node from an object ID.
func (query *Query) loadByOID(idx nodeIndex) error {
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

	ty, content, err := query.store.ReadBytesContent(id)
	if err != nil {
		if stderrors.Is(err, objectstore.ErrObjectNotFound) {
			return &giterrors.ObjectMissingError{OID: id}
		}

		return err
	}

	if ty != objecttype.TypeCommit {
		return &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
	}

	commitObj, err := object.ParseCommit(content, id.Algorithm())
	if err != nil {
		return err
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
