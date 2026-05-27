package reachability

import (
	"fmt"

	"lindenii.org/go/furgit/errors"
	objecttree "lindenii.org/go/furgit/object/tree"
	objecttype "lindenii.org/go/furgit/object/type"
)

func (walk *Walk) expandObjects(item walkItem) ([]walkItem, error) {
	ty, _, err := walk.reachability.fetcher.Header(item.id)
	if err != nil {
		return nil, err
	}

	if item.want != objecttype.TypeInvalid && ty != item.want {
		return nil, &errors.ObjectTypeError{OID: item.id, Got: ty, Want: item.want}
	}

	switch ty {
	case objecttype.TypeBlob:
		return nil, nil
	case objecttype.TypeCommit:
		commit, err := walk.reachability.fetcher.ExactCommit(item.id)
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(commit.Object().Parents)+1)

		next = append(next, walkItem{id: commit.Object().Tree, want: objecttype.TypeTree})
		for _, parent := range commit.Object().Parents {
			next = append(next, walkItem{id: parent, want: objecttype.TypeCommit})
		}

		return next, nil
	case objecttype.TypeTree:
		tree, err := walk.reachability.fetcher.ExactTree(item.id)
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(tree.Object().Entries))
		for _, entry := range tree.Object().Entries {
			switch entry.Mode {
			case objecttree.FileModeGitlink:
				continue
			case objecttree.FileModeDir:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeTree})
			case objecttree.FileModeRegular, objecttree.FileModeExecutable, objecttree.FileModeSymlink:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeBlob})
			}
		}

		return next, nil
	case objecttype.TypeTag:
		tag, err := walk.reachability.fetcher.ExactTag(item.id)
		if err != nil {
			return nil, err
		}

		return []walkItem{{id: tag.Object().TargetID, want: tag.Object().TargetType}}, nil
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &errors.ObjectTypeError{OID: item.id, Got: ty, Want: item.want}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}
