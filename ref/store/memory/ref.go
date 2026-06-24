package memory

import "lindenii.org/go/furgit/object/id"

// storedRef is the internal representation of one reference.
//
// Unlike the public ref values,
// it carries no name of its own;
// the name is the map key.
type storedRef struct {
	kind   storedKind
	id     id.ObjectID  //exhaustruct:optional
	target string       //exhaustruct:optional
	peeled *id.ObjectID //exhaustruct:optional
}

type storedKind uint8

const (
	storedMissing storedKind = iota
	storedDirect
	storedSymbolic
)

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
