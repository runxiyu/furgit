package read

import "codeberg.org/lindenii/furgit/objectid"

// Reader provides read-only access to one mmap-backed commit-graph snapshot.
//
// It is safe for concurrent read-only queries.
// Values returned by Reader methods are only valid until the reader is closed
// when explicitly documented on that method.
type Reader struct {
	algo        objectid.Algorithm
	hashVersion uint8

	layers []layer
	total  uint32
}
