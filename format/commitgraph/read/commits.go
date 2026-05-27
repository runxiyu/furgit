package read

import objectid "lindenii.org/go/furgit/object/id"

// Commit stores decoded commit-graph record data.
type Commit struct {
	OID            objectid.ObjectID
	TreeOID        objectid.ObjectID
	Parent1        ParentRef
	Parent2        ParentRef
	ExtraParents   []Position
	CommitTimeUnix int64
	GenerationV1   uint32
	GenerationV2   uint64
}

// NumCommits returns total commits across loaded layers.
func (reader *Reader) NumCommits() uint32 {
	return reader.total
}
