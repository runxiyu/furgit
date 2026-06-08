package fetch

import (
	"fmt"

	"lindenii.org/go/furgit/object"
	oid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/stored"
)

// ExactObject reads, parses, and wraps the object at id without constraining
// its concrete object kind.
//
// Labels: Life-Parent.
func (fetcher *Fetcher) ExactObject(id oid.ObjectID) (*stored.Stored[object.Object], error) {
	parsed, err := fetcher.parseObject(id)
	if err != nil {
		return nil, err
	}

	return stored.New(id, parsed), nil
}

func (fetcher *Fetcher) parseObject(id oid.ObjectID) (object.Object, error) {
	ty, content, err := fetcher.store.ReadBytesContent(id)
	if err != nil {
		return nil, wrapObjectReadError(id, err)
	}

	parsed, err := object.ParseWithoutHeader(ty, content, id.ObjectFormat())
	if err != nil {
		tyName, ok := ty.Name()
		if !ok {
			tyName = fmt.Sprintf("type %d", ty)
		}

		return nil, fmt.Errorf("object/fetch: parse object %s (%s): %w", id, tyName, err)
	}

	return parsed, nil
}
