package resolve

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func (r *Resolver) parseObject(id objectid.ObjectID) (object.Object, error) {
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

		return nil, fmt.Errorf("object/resolve: parse object %s (%s): %w", id, tyName, err)
	}

	return parsed, nil
}
