package dual

import objectstore "codeberg.org/lindenii/furgit/object/store"

// BeginQuarantine creates one coordinated dual quarantine spanning both stores.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginQuarantine(opts objectstore.QuarantineOptions) (objectstore.WriterQuarantine, error) {
	return dual.beginQuarantine(opts)
}

// BeginObjectQuarantine creates one coordinated dual quarantine spanning both
// stores and returns it as an object-wise quarantine.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginObjectQuarantine(opts objectstore.ObjectQuarantineOptions) (objectstore.ObjectQuarantine, error) {
	quarantine, err := dual.beginQuarantine(objectstore.QuarantineOptions{
		Object: opts,
	})
	if err != nil {
		return nil, err
	}

	return quarantine, nil
}

// BeginPackQuarantine creates one coordinated dual quarantine spanning both
// stores and returns it as a pack-wise quarantine.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginPackQuarantine(opts objectstore.PackQuarantineOptions) (objectstore.PackQuarantine, error) {
	quarantine, err := dual.beginQuarantine(objectstore.QuarantineOptions{
		Pack: opts,
	})
	if err != nil {
		return nil, err
	}

	return quarantine, nil
}

func (dual *Dual) beginQuarantine(opts objectstore.QuarantineOptions) (*quarantine, error) {
	objectQ, err := dual.object.BeginObjectQuarantine(opts.Object)
	if err != nil {
		return nil, err
	}

	packQ, err := dual.pack.BeginPackQuarantine(opts.Pack)
	if err != nil {
		_ = objectQ.Discard()

		return nil, err
	}

	return newQuarantine(objectQ, packQ), nil
}
