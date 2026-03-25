package memory

import objecttype "codeberg.org/lindenii/furgit/object/type"

// storedObject is one in-memory object entry.
type storedObject struct {
	ty      objecttype.Type
	content []byte
}
