package commitquery

import (
	stderrors "errors"

	giterrors "codeberg.org/lindenii/furgit/errors"
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectcommit "codeberg.org/lindenii/furgit/object/commit"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ensureLoaded completes one node's metadata load if it has not been loaded yet.
func (query *query) ensureLoaded(idx nodeIndex) error {
	if query.nodes[idx].loaded {
		return nil
	}

	if query.nodes[idx].hasGraphPos {
		return query.loadByGraphPos(idx)
	}

	return query.loadByOID(idx)
}

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

	commitObj, err := objectcommit.Parse(content, id.Algorithm())
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
