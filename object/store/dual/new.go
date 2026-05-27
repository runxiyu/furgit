package dual

import (
	objectstore "lindenii.org/go/furgit/object/store"
	objectmix "lindenii.org/go/furgit/object/store/mix"
)

// New creates one dual object store from borrowed object-wise and pack-wise
// stores.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(
	object interface {
		objectstore.Reader
		objectstore.ObjectWriter
		objectstore.ObjectQuarantiner
	},
	pack interface {
		objectstore.Reader
		objectstore.PackWriter
		objectstore.PackQuarantiner
	},
) *Dual {
	return &Dual{
		object: object,
		pack:   pack,
		reader: objectmix.New(object, pack),
	}
}
