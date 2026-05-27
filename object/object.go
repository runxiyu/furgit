package object

import "lindenii.org/go/furgit/object/typ"

// Object is a Git object.
type Object interface {
	ObjectType() typ.Type
	AppendWithoutHeader(dst []byte) ([]byte, error)
	AppendWithHeader(dst []byte) ([]byte, error)
}
