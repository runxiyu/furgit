package reachability

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func (r *Reachability) peelRootToDomain(id objectid.ObjectID, domain Domain) (objectid.ObjectID, error) {
	if err := validateDomain(domain); err != nil {
		return objectid.ObjectID{}, err
	}
	for {
		ty, err := r.readHeaderType(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}
		if ty != objecttype.TypeTag {
			if domain == DomainCommits && ty != objecttype.TypeCommit {
				return objectid.ObjectID{}, &ErrObjectType{OID: id, Got: ty, Want: objecttype.TypeCommit}
			}
			return id, nil
		}

		content, err := r.readBytesContent(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}
		tag, err := object.ParseTag(content, id.Algorithm())
		if err != nil {
			return objectid.ObjectID{}, err
		}
		id = tag.Target
	}
}
