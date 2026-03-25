// Package object parses and serializes objects such as blob, tree, commit, and
// tag.
package object

import objecttype "codeberg.org/lindenii/furgit/object/type"

// Object is a Git object that can serialize itself.
type Object interface {
	ObjectType() objecttype.Type
	SerializeWithoutHeader() ([]byte, error)
	SerializeWithHeader() ([]byte, error)
}
