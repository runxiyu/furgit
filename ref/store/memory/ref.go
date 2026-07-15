package memory

import (
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
)

// storedRef is the internal representation of one reference.
//
// Unlike the public ref values,
// it carries no name of its own;
// the name is the map key.
type storedRef struct {
	kind      storedKind
	id        id.ObjectID   //exhaustruct:optional
	target    string        //exhaustruct:optional
	peelState ref.PeelState //exhaustruct:optional
	peeledID  id.ObjectID   //exhaustruct:optional
}

type storedKind uint8

const (
	storedMissing storedKind = iota
	storedDirect
	storedSymbolic
)
