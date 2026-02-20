// Package object provides Git object models and codecs.
package object

import "codeberg.org/lindenii/furgit/objecttype"

// Object is a Git object that can serialize itself.
type Object interface {
	ObjectType() objecttype.Type
	SerializeWithoutHeader() ([]byte, error)
	SerializeWithHeader() ([]byte, error)
}
