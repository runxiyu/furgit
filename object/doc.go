// Package object provides the shared [Object] interface
// and parsing functions for Git object values.
//
// Concrete object forms such as
// [blob], [tree], [commit], and [tag]
// live in subpackages.
// Use [codeberg.org/lindenii/furgit/object/stored] when object values
// need to be paired with the object IDs
// they were loaded under.
//
// This package also encapsulates other object-related packages,
// including but not limited to
// object IDs, object types, and object stores.
package object
