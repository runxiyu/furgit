package memory

import objectid "codeberg.org/lindenii/furgit/object/id"

// Because the public one includes the ref's name/identity.

type storedKind uint8

const (
	storedMissing storedKind = iota
	storedDetached
	storedSymbolic
)

// Missing is obviously not the best design
// but it does make it easier to operate on internally.
// Might make a tagged union wrapper, though...
// Or might just make a wrapper struct that has an "ok" bool.

type storedRef struct {
	kind   storedKind
	id     objectid.ObjectID
	target string
	peeled *objectid.ObjectID
}

func cloneStoredRef(stored storedRef) storedRef {
	if stored.peeled == nil {
		return stored
	}

	peeled := *stored.peeled
	stored.peeled = &peeled

	return stored
}

func cloneRefs(refs map[string]storedRef) map[string]storedRef {
	cloned := make(map[string]storedRef, len(refs))
	for name, stored := range refs {
		cloned[name] = cloneStoredRef(stored)
	}

	return cloned
}
