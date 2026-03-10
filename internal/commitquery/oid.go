package commitquery

import (
	stderrors "errors"

	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ID returns the canonical object ID of one internal node.
func (ctx *Context) ID(idx NodeIndex) objectid.ObjectID {
	return ctx.nodes[idx].id
}

// CommitTime returns the committer timestamp used for one internal node.
func (ctx *Context) CommitTime(idx NodeIndex) int64 {
	return ctx.nodes[idx].commitTime
}

// ResolveOID resolves one commit object ID to one internal query node.
func (ctx *Context) ResolveOID(id objectid.ObjectID) (NodeIndex, error) {
	idx, ok := ctx.byOID[id]
	if ok {
		err := ctx.ensureLoaded(idx)
		if err != nil {
			return 0, err
		}

		return idx, nil
	}

	idx = ctx.newNode(id)
	ctx.byOID[id] = idx

	err := ctx.loadByOID(idx)
	if err != nil {
		delete(ctx.byOID, id)

		return 0, err
	}

	return idx, nil
}

// loadByOID populates one node from an object ID.
func (ctx *Context) loadByOID(idx NodeIndex) error {
	id := ctx.nodes[idx].id

	if ctx.graph != nil {
		pos, err := ctx.graph.Lookup(id)
		if err != nil {
			if _, ok := stderrors.AsType[*commitgraphread.NotFoundError](err); !ok {
				return err
			}
		} else {
			return ctx.loadCommitAtGraphPos(idx, pos)
		}
	}

	ty, content, err := ctx.store.ReadBytesContent(id)
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

	parents := make([]Parent, 0, len(commitObj.Parents))
	for _, parentID := range commitObj.Parents {
		parents = append(parents, Parent{ID: parentID})
	}

	commit := Commit{
		ID:         id,
		Parents:    parents,
		CommitTime: commitObj.Committer.WhenUnix,
	}

	return ctx.populateNode(idx, commit)
}
