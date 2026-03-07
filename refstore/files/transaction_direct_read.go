package files

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/refstore"
)

func (tx *Transaction) directRead(name string) (directRef, error) {
	loc := tx.store.loosePath(name)
	hasPacked := false

	if loc.root == rootCommon && refname.ParseWorktree(name).Type == refname.WorktreeShared {
		packed, packedErr := tx.store.readPackedRefs()
		if packedErr != nil {
			return directRef{}, packedErr
		}

		_, hasPacked = packed.byName[name]
	}

	loose, err := tx.store.readLooseRef(name)
	if err == nil {
		switch loose := loose.(type) {
		case ref.Detached:
			return directRef{
				kind:     directDetached,
				name:     name,
				id:       loose.ID,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		case ref.Symbolic:
			return directRef{
				kind:     directSymbolic,
				name:     name,
				target:   loose.Target,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		default:
			return directRef{}, fmt.Errorf("refstore/files: unsupported reference type %T", loose)
		}
	}

	if !errors.Is(err, refstore.ErrReferenceNotFound) {
		info, statErr := tx.store.rootFor(loc.root).Stat(loc.path)
		if statErr != nil || !info.IsDir() {
			return directRef{}, err
		}
	}

	if hasPacked {
		packed, packedErr := tx.store.readPackedRefs()
		if packedErr != nil {
			return directRef{}, packedErr
		}

		detached := packed.byName[name]

		return directRef{
			kind:     directDetached,
			name:     name,
			id:       detached.ID,
			isPacked: true,
		}, nil
	}

	return directRef{
		kind: directMissing,
		name: name,
	}, nil
}
