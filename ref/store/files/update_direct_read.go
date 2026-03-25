package files

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/ref/refname"
	"codeberg.org/lindenii/furgit/ref/store"
)

func (executor *refUpdateExecutor) directRead(name string) (directRefState, error) {
	loc := executor.store.loosePath(name)
	hasPacked := false

	if loc.root == rootCommon && refname.ParseWorktree(name).Type == refname.WorktreeShared {
		packed, packedErr := executor.store.readPackedRefs()
		if packedErr != nil {
			return directRefState{}, packedErr
		}

		_, hasPacked = packed.byName[name]
	}

	loose, err := executor.store.readLooseRef(name)
	if err == nil {
		switch loose := loose.(type) {
		case ref.Detached:
			return directRefState{
				kind:     directDetached,
				name:     name,
				id:       loose.ID,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		case ref.Symbolic:
			return directRefState{
				kind:     directSymbolic,
				name:     name,
				target:   loose.Target,
				isLoose:  true,
				isPacked: hasPacked,
			}, nil
		default:
			return directRefState{}, fmt.Errorf("refstore/files: unsupported reference type %T", loose)
		}
	}

	if !errors.Is(err, refstore.ErrReferenceNotFound) {
		info, statErr := executor.store.rootFor(loc.root).Stat(loc.path)
		if statErr != nil || !info.IsDir() {
			return directRefState{}, err
		}
	}

	if hasPacked {
		packed, packedErr := executor.store.readPackedRefs()
		if packedErr != nil {
			return directRefState{}, packedErr
		}

		detached := packed.byName[name]

		return directRefState{
			kind:     directDetached,
			name:     name,
			id:       detached.ID,
			isPacked: true,
		}, nil
	}

	return directRefState{
		kind: directMissing,
		name: name,
	}, nil
}
