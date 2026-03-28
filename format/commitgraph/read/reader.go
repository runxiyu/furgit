package read

import objectid "codeberg.org/lindenii/furgit/object/id"

// Reader provides read-only access to one mmap-backed commit-graph snapshot.
//
// Labels: MT-ReadSafe, Close-Caller.
type Reader struct {
	algo        objectid.Algorithm
	hashVersion uint8

	layers []layer
	total  uint32
}
