package dual

import objectstore "codeberg.org/lindenii/furgit/object/store"

type objectSide interface {
	objectstore.Reader
	objectstore.ObjectWriter
	objectstore.ObjectQuarantiner
}

type packSide interface {
	objectstore.Reader
	objectstore.PackWriter
	objectstore.PackQuarantiner
}

// Dual composes one object-wise store and one pack-wise store into one logical
// object store.
//
// Reads are served from the combined object reader of both stores. Individual
// object writes are routed to the object-wise store, and pack writes are routed
// to the pack-wise store. Coordinated quarantines go across both stores.
type Dual struct {
	object objectSide
	pack   packSide
	reader objectstore.Reader
}

var (
	_ objectstore.Reader      = (*Dual)(nil)
	_ objectstore.Writer      = (*Dual)(nil)
	_ objectstore.Quarantiner = (*Dual)(nil)
)
