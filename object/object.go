// Package object provides shared object interfaces.
package object

import objecttype "codeberg.org/lindenii/furgit/object/type"

// Object is a Git object.
type Object interface {
	ObjectType() objecttype.Type
	SerializeWithoutHeader() ([]byte, error)
	SerializeWithHeader() ([]byte, error)
}
