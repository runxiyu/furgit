package files

import (
	"errors"

	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Resolve resolves one reference name from the files store visible namespace.
func (store *Store) Resolve(name string) (ref.Ref, error) { //nolint:ireturn
	if name == "" {
		return nil, refstore.ErrReferenceNotFound
	}

	resolved, err := store.readLooseRef(name)
	if err == nil {
		return resolved, nil
	}

	if !errors.Is(err, refstore.ErrReferenceNotFound) {
		refPath := store.loosePath(name)

		info, statErr := store.rootFor(refPath.root).Stat(refPath.path)
		if statErr != nil || !info.IsDir() {
			return nil, err
		}
	}

	packed, packedErr := store.readPackedRefs()
	if packedErr != nil {
		return nil, packedErr
	}

	detached, ok := packed.byName[name]
	if !ok {
		return nil, refstore.ErrReferenceNotFound
	}

	return detached, nil
}
