package memory

import "codeberg.org/lindenii/furgit/objecttype"

// storedObject is one in-memory object entry.
type storedObject struct {
	ty      objecttype.Type
	content []byte
}
