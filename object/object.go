package object

import "codeberg.org/lindenii/furgit/object/typ"

// Object is a Git object.
type Object interface {
	ObjectType() typ.Type
	AppendWithoutHeader([]byte) ([]byte, error)
	AppendWithHeader([]byte) ([]byte, error)
}
