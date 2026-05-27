package memory

import objecttype "lindenii.org/go/furgit/object/type"

// storedObject is one in-memory object entry.
type storedObject struct {
	ty      objecttype.Type
	content []byte
}
