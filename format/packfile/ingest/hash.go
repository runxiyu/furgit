package ingest

import (
	"fmt"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// hashCanonicalObject hashes canonical object bytes (header+content).
func hashCanonicalObject(algo objectid.Algorithm, ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	header, ok := objectheader.Encode(ty, int64(len(content)))
	if !ok {
		return objectid.ObjectID{}, fmt.Errorf("packfile/ingest: encode object header for type %d", ty)
	}

	hashImpl, err := algo.New()
	if err != nil {
		return objectid.ObjectID{}, err
	}

	_, _ = hashImpl.Write(header)
	_, _ = hashImpl.Write(content)

	return objectid.FromBytes(algo, hashImpl.Sum(nil))
}
