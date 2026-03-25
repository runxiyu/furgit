package fetch

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (r *Fetcher) parseObject(id objectid.ObjectID) (object.Object, error) {
	ty, content, err := r.store.ReadBytesContent(id)
	if err != nil {
		return nil, err
	}

	parsed, err := object.ParseObjectWithoutHeader(ty, content, id.Algorithm())
	if err != nil {
		tyName, ok := objecttype.Name(ty)
		if !ok {
			tyName = fmt.Sprintf("type %d", ty)
		}

		return nil, fmt.Errorf("object/fetch: parse object %s (%s): %w", id, tyName, err)
	}

	return parsed, nil
}
