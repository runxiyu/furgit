package commitgraph

import "codeberg.org/lindenii/furgit/objectid"

// Reader provides read-only access to one mmap-backed commit-graph snapshot.
//
// It is safe for concurrent read-only queries.
type Reader struct {
	algo        objectid.Algorithm
	hashVersion uint8

	layers []layer
	total  uint32
}
