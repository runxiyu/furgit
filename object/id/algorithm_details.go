package objectid

import "hash"

type algorithmDetails struct {
	name       string
	size       int
	packHashID uint32
	sum        func([]byte) ObjectID
	new        func() hash.Hash
	emptyTree  ObjectID
}

func (algo Algorithm) info() algorithmDetails {
	return algorithmTable[algo]
}
