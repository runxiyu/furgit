package reachability

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objecttype"
)

func (walk *Walk) expandObjects(item walkItem) ([]walkItem, error) {
	ty, err := walk.readHeaderType(item.id)
	if err != nil {
		return nil, err
	}

	if item.want != objecttype.TypeInvalid && ty != item.want {
		return nil, &ObjectTypeError{OID: item.id, Got: ty, Want: item.want}
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
		return nil, &ObjectTypeError{OID: item.id, Got: ty, Want: item.want}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}
