package dual

import (
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/mix"
)

// objectSide is the object-wise half of a [Dual].
type objectSide interface {
	store.ObjectReader
	store.ObjectWriter
	store.ObjectQuarantiner
}

// packSide is the pack-wise half of a [Dual].
type packSide interface {
	store.ObjectReader
	store.PackWriter
	store.PackQuarantiner
}

// Dual composes one object-wise store and one pack-wise store
// into one logical object store.
//
// Reads are served from a most-recently-used view over both sides,
// object-wise writes are routed to the object side,
// pack-wise writes are routed to the pack side,
// and quarantines span the sides they cover.
//
// Labels: Deps-Borrowed, Life-Parent, Close-Caller.
type Dual struct {
	object objectSide
	pack   packSide
	reader store.ObjectReader
}

var (
	_ store.ObjectReader           = (*Dual)(nil)
	_ store.ObjectWriter           = (*Dual)(nil)
	_ store.PackWriter             = (*Dual)(nil)
	_ store.ObjectQuarantiner      = (*Dual)(nil)
	_ store.PackQuarantiner        = (*Dual)(nil)
	_ store.CoordinatedQuarantiner = (*Dual)(nil)
)

// New composes object-wise and pack-wise stores into one logical store.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(
	object objectSide,
	pack packSide,
) *Dual {
	return &Dual{
		object: object,
		pack:   pack,
		reader: mix.New(object, pack),
	}
}
