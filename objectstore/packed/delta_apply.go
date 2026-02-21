package packed

import (
	"fmt"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaResolveContent resolves one object's content bytes from its pack location.
func (store *Store) deltaResolveContent(start location) (objecttype.Type, []byte, error) {
	plan, err := store.deltaPlanFor(start)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	baseType, out, err := store.deltaResolveBase(plan)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}
	for i := len(plan.frames) - 1; i >= 0; i-- {
		frame := plan.frames[i]
		pack, err := store.openPack(frame.packName)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
		delta, err := inflateAt(pack, frame.dataOffset, -1)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
		out, err = deltaapply.Apply(out, delta)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
	}
	if int64(len(out)) != plan.declaredSize {
		return objecttype.TypeInvalid, nil, fmt.Errorf(
			"objectstore/packed: resolved content size mismatch: got %d want %d",
			len(out),
			plan.declaredSize,
		)
	}
	return baseType, out, nil
}
