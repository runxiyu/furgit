// Package tag provides representations, parsers, and serializers for tag objects.
package tag

import (
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectsignature "codeberg.org/lindenii/furgit/object/signature"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// Tag represents a Git annotated tag object.
type Tag struct {
	Target     objectid.ObjectID
	TargetType objecttype.Type
	Name       []byte
	Tagger     *objectsignature.Signature
	Message    []byte
}
