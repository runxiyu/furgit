package reachability

import (
	"errors"
	"fmt"

	giterrors "codeberg.org/lindenii/furgit/errors"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func validateDomain(domain Domain) error {
	switch domain {
	case DomainCommits, DomainObjects:
		return nil
	default:
		return fmt.Errorf("reachability: invalid domain %d", domain)
	}
}

func containsOID(set map[objectid.ObjectID]struct{}, id objectid.ObjectID) bool {
	if len(set) == 0 {
		return false
	}

	_, ok := set[id]

	return ok
}

// The following helpers exist because we don't have unified error handling across the entire project.
// This will be fixed later.

func (walk *Walk) readHeaderType(id objectid.ObjectID) (objecttype.Type, error) {
	return walk.reachability.readHeaderType(id)
}

func (r *Reachability) readHeaderType(id objectid.ObjectID) (objecttype.Type, error) {
	ty, _, err := r.store.ReadHeader(id)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return objecttype.TypeInvalid, &giterrors.ObjectMissingError{OID: id}
		}

		return objecttype.TypeInvalid, err
	}

	return ty, nil
}

func (walk *Walk) readBytesContent(id objectid.ObjectID) ([]byte, error) {
	content, err := walk.reachability.readBytesContent(id)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (r *Reachability) readBytesContent(id objectid.ObjectID) ([]byte, error) {
	_, content, err := r.store.ReadBytesContent(id)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return nil, &giterrors.ObjectMissingError{OID: id}
		}

		return nil, err
	}

	return content, nil
}
