// Package object provides the shared [Object] interface and parsing functions
// for Git object values.
//
// Concrete object forms such as [blob], [tree], [commit], and [tag] live in
// subpackages. Use [lindenii.org/go/furgit/object/stored] when object
// values need to be paired with the object IDs they were loaded under.
package object
