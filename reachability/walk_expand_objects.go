package reachability

import (
	"fmt"

	"codeberg.org/lindenii/furgit/errors"
	objectcommit "codeberg.org/lindenii/furgit/object/commit"
	objecttag "codeberg.org/lindenii/furgit/object/tag"
	objecttree "codeberg.org/lindenii/furgit/object/tree"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (walk *Walk) expandObjects(item walkItem) ([]walkItem, error) {
	ty, err := walk.readHeaderType(item.id)
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
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		commit, err := objectcommit.Parse(content, item.id.Algorithm())
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

		tree, err := objecttree.Parse(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		next := make([]walkItem, 0, len(tree.Entries))
		for _, entry := range tree.Entries {
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
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}

		tag, err := objecttag.Parse(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		return []walkItem{{id: tag.Target, want: tag.TargetType}}, nil
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &errors.ObjectTypeError{OID: item.id, Got: ty, Want: item.want}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}
