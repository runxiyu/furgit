package reachability

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ErrObjectMissing indicates that a referenced object is absent from the store.
type ErrObjectMissing struct {
	OID objectid.ObjectID
}

func (e *ErrObjectMissing) Error() string {
	return fmt.Sprintf("reachability: missing object %s", e.OID)
}

// ErrObjectType indicates that a referenced object has a different type than
// what traversal expected on that edge.
type ErrObjectType struct {
	OID  objectid.ObjectID
	Got  objecttype.Type
	Want objecttype.Type
}

func (e *ErrObjectType) Error() string {
	gotName, gotOK := objecttype.Name(e.Got)
	if !gotOK {
		gotName = fmt.Sprintf("type(%d)", e.Got)
	}

	wantName, wantOK := objecttype.Name(e.Want)
	if !wantOK {
		wantName = fmt.Sprintf("type(%d)", e.Want)
	}

	return fmt.Sprintf("reachability: object %s has type %s, want %s", e.OID, gotName, wantName)
}
