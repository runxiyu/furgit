package dual

import objectstore "lindenii.org/go/furgit/object/store"

// BeginQuarantine creates one coordinated dual quarantine spanning both stores.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginQuarantine(opts objectstore.QuarantineOptions) (objectstore.Quarantine, error) {
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
