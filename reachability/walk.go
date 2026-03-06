package reachability

import (
	"errors"
	"fmt"
	"iter"

	"codeberg.org/lindenii/furgit/format/commitgraph"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Walk is one single-use iterator-style traversal.
type Walk struct {
	reachability *Reachability
	domain       Domain
	haves        map[objectid.ObjectID]struct{}
	wants        map[objectid.ObjectID]struct{}
	strict       bool

	seqUsed bool
	err     error
}

// Seq returns the traversal sequence. It is single-use.
func (walk *Walk) Seq() iter.Seq[objectid.ObjectID] {
	if walk.seqUsed {
		return func(yield func(objectid.ObjectID) bool) {
			_ = yield

			if walk.err == nil {
				walk.err = errors.New("reachability: walk sequence already consumed")
			}
		}
	}

	walk.seqUsed = true

	return func(yield func(objectid.ObjectID) bool) {
		if walk.err != nil {
			return
		}

		stack := walk.initialStack()

		var err error

		visited := make(map[objectid.ObjectID]struct{}, len(stack))
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if containsOID(walk.haves, item.id) {
				continue
			}

			if _, ok := visited[item.id]; ok {
				continue
			}

			visited[item.id] = struct{}{}

			var next []walkItem

			next, err = walk.expand(item)
			if err != nil {
				walk.err = err

				return
			}

			if !yield(item.id) {
				return
			}

			stack = append(stack, next...)
		}
	}
}

// Err returns the terminal error, if any, once Seq has been consumed.
func (walk *Walk) Err() error {
	return walk.err
}

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}

func (walk *Walk) initialStack() []walkItem {
	if len(walk.wants) == 0 {
		return nil
	}

	stack := make([]walkItem, 0, len(walk.wants))
	for want := range walk.wants {
		stack = append(stack, walkItem{id: want, want: objecttype.TypeInvalid})
	}

	return stack
}

func (walk *Walk) expand(item walkItem) ([]walkItem, error) {
	if walk.domain == DomainCommits {
		return walk.expandCommits(item)
	}

	return walk.expandObjects(item)
}

func (walk *Walk) expandCommits(item walkItem) ([]walkItem, error) {
	if walk.reachability.graph != nil {
		next, graphUsed, err := walk.expandCommitsFromGraph(item.id)
		if err != nil {
			return nil, err
		}

		if graphUsed {
			if walk.strict {
				err := walk.validateCommitObject(item.id)
				if err != nil {
					return nil, err
				}
			}

			return next, nil
		}
	}

	ty, err := walk.readHeaderType(item.id)
	if err != nil {
		return nil, err
	}

	switch ty {
	case objecttype.TypeCommit:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		commit, err := object.ParseCommit(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(commit.Parents))
		for _, parent := range commit.Parents {
			next = append(next, walkItem{id: parent, want: objecttype.TypeInvalid})
		}

		return next, nil
	case objecttype.TypeTag:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		tag, err := object.ParseTag(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		return []walkItem{{id: tag.Target, want: objecttype.TypeInvalid}}, nil
	case objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeInvalid,
		objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: objecttype.TypeCommit}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}

func (walk *Walk) validateCommitObject(id objectid.ObjectID) error {
	ty, err := walk.readHeaderType(id)
	if err != nil {
		return err
	}

	if ty != objecttype.TypeCommit {
		return &ErrObjectType{OID: id, Got: ty, Want: objecttype.TypeCommit}
	}

	content, err := walk.readBytesContent(id)
	if err != nil {
		return err
	}

	_, err = object.ParseCommit(content, id.Algorithm())

	return err
}

func (walk *Walk) expandCommitsFromGraph(id objectid.ObjectID) ([]walkItem, bool, error) {
	pos, err := walk.reachability.graph.Lookup(id)
	if err != nil {
		var notFound *commitgraph.ErrNotFound
		if errors.As(err, &notFound) {
			return nil, false, nil
		}

		return nil, true, err
	}

	commit, err := walk.reachability.graph.CommitAt(pos)
	if err != nil {
		return nil, true, err
	}

	next := make([]walkItem, 0, 2+len(commit.ExtraParents))

	if commit.Parent1.Valid {
		parentOID, err := walk.reachability.graph.OIDAt(commit.Parent1.Pos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	if commit.Parent2.Valid {
		parentOID, err := walk.reachability.graph.OIDAt(commit.Parent2.Pos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	for _, parentPos := range commit.ExtraParents {
		parentOID, err := walk.reachability.graph.OIDAt(parentPos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	return next, true, nil
}

func (walk *Walk) expandObjects(item walkItem) ([]walkItem, error) {
	ty, err := walk.readHeaderType(item.id)
	if err != nil {
		return nil, err
	}

	if item.want != objecttype.TypeInvalid && ty != item.want {
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: item.want}
	}

	switch ty {
	case objecttype.TypeBlob:
		return nil, nil
	case objecttype.TypeCommit:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		commit, err := object.ParseCommit(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(commit.Parents)+1)

		next = append(next, walkItem{id: commit.Tree, want: objecttype.TypeTree})
		for _, parent := range commit.Parents {
			next = append(next, walkItem{id: parent, want: objecttype.TypeCommit})
		}

		return next, nil
	case objecttype.TypeTree:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		tree, err := object.ParseTree(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(tree.Entries))
		for _, entry := range tree.Entries {
			switch entry.Mode {
			case object.FileModeGitlink:
				continue
			case object.FileModeDir:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeTree})
			case object.FileModeRegular, object.FileModeExecutable, object.FileModeSymlink:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeBlob})
			}
		}

		return next, nil
	case objecttype.TypeTag:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		tag, err := object.ParseTag(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		return []walkItem{{id: tag.Target, want: tag.TargetType}}, nil
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: item.want}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}
