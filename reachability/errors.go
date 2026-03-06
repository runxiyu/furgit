package reachability

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ObjectMissingError indicates that a referenced object is absent from the store.
type ObjectMissingError struct {
	OID objectid.ObjectID
}

func (e *ObjectMissingError) Error() string {
	return fmt.Sprintf("reachability: missing object %s", e.OID)
}

// ObjectTypeError indicates that a referenced object has a different type than
// what traversal expected on that edge.
type ObjectTypeError struct {
	OID  objectid.ObjectID
	Got  objecttype.Type
	Want objecttype.Type
}

func (e *ObjectTypeError) Error() string {
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
