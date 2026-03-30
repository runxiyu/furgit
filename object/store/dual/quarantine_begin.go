package dual

import objectstore "codeberg.org/lindenii/furgit/object/store"

// TODO: This doesn't actually make sense. We need a combined quarantine.

// BeginObjectQuarantine creates one coordinated dual quarantine spanning both
// stores and returns it as an object-wise quarantine.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginObjectQuarantine(_ objectstore.ObjectQuarantineOptions) (objectstore.ObjectQuarantine, error) {
	quarantine, err := dual.beginQuarantine()
	if err != nil {
		return nil, err
	}

	return quarantine, nil
}

// BeginPackQuarantine creates one coordinated dual quarantine spanning both
// stores and returns it as a pack-wise quarantine.
//
// Labels: Deps-Borrowed, Life-Parent, Close-No.
func (dual *Dual) BeginPackQuarantine(_ objectstore.PackQuarantineOptions) (objectstore.PackQuarantine, error) {
	quarantine, err := dual.beginQuarantine()
	if err != nil {
		return nil, err
	}

	return quarantine, nil
}

func (dual *Dual) beginQuarantine() (*quarantine, error) {
	objectQ, err := dual.object.BeginObjectQuarantine(objectstore.ObjectQuarantineOptions{})
	if err != nil {
		return nil, err
	}

	packQ, err := dual.pack.BeginPackQuarantine(objectstore.PackQuarantineOptions{})
	if err != nil {
		_ = objectQ.Discard()

		return nil, err
	}

	return newQuarantine(objectQ, packQ), nil
}
