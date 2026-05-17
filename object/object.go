package object

import objecttype "codeberg.org/lindenii/furgit/object/type"

// Object is a Git object.
type Object interface {
	ObjectType() objecttype.Type
	BytesWithoutHeader() ([]byte, error)
	BytesWithHeader() ([]byte, error)
}
