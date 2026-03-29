package fetch

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

func (r *Fetcher) parseObject(id objectid.ObjectID) (object.Object, error) {
	ty, content, err := r.store.ReadBytesContent(id)
	if err != nil {
		return nil, wrapObjectReadError(id, err)
	}

	parsed, err := object.ParseWithoutHeader(ty, content, id.Algorithm())
	if err != nil {
		tyName, ok := ty.Name()
		if !ok {
			tyName = fmt.Sprintf("type %d", ty)
		}

		return nil, fmt.Errorf("object/fetch: parse object %s (%s): %w", id, tyName, err)
	}

	return parsed, nil
}
