package reachability

import (
	"fmt"

	"codeberg.org/lindenii/furgit/errors"
	objectcommit "codeberg.org/lindenii/furgit/object/commit"
	objecttag "codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (walk *Walk) expandCommits(item walkItem) ([]walkItem, error) {
	if walk.reachability.graph != nil { //nolint:nestif
		next, graphUsed, err := walk.expandCommitsFromGraph(item.id)
		if err != nil {
			return nil, err
		}

		if graphUsed && walk.strict {
			err = walk.validateCommitObject(item.id)
			if err != nil {
				return nil, err
			}
		}

		if graphUsed {
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

		commit, err := objectcommit.Parse(content, item.id.Algorithm())
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

		tag, err := objecttag.Parse(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}

		return []walkItem{{id: tag.Target, want: objecttype.TypeInvalid}}, nil
	case objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeInvalid,
		objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &errors.ObjectTypeError{OID: item.id, Got: ty, Want: objecttype.TypeCommit}
	}

	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}
