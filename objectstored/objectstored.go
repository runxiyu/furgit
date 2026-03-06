package objectstored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// StoredTag is a parsed tag paired with its storage ID.
type StoredTag struct {
	id  objectid.ObjectID
	tag *object.Tag
}

// NewStoredTag creates one stored tag wrapper.
func NewStoredTag(id objectid.ObjectID, tag *object.Tag) *StoredTag {
	return &StoredTag{id: id, tag: tag}
}

// ID returns the object ID this tag was loaded from.
func (stored *StoredTag) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed tag as the generic object interface.
func (stored *StoredTag) Object() object.Object { //nolint:ireturn
	return stored.tag
}

// Tag returns the parsed tag value.
func (stored *StoredTag) Tag() *object.Tag {
	return stored.tag
}
