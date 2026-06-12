package packed

import (
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

// lookup finds the pack containing objectID
// and the entry offset within it,
// probing packs in most-recently-used-ish order.
//
// Labels: Life-Parent.
func (packed *Packed) lookup(objectID id.ObjectID) (*pack, uint64, error) {
	if objectID.ObjectFormat() != packed.objectFormat {
		return nil, 0, fmt.Errorf(
			"%w: got %s want %s",
			id.ErrInvalidObjectFormat, objectID.ObjectFormat(), packed.objectFormat,
		)
	}

	oid := objectID.RawBytes()

	for _, p := range packed.order.Keys() {
		offset, found, err := p.idx.Lookup(oid)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
		}

		if !found {
			continue
		}

		packed.order.Touch(p)

		return p, offset, nil
	}

	return nil, 0, store.ErrObjectNotFound
}
