package ref

import objectid "lindenii.org/go/furgit/object/id"

// PeelState describes a direct reference's cached peel knowledge.
type PeelState uint8

const (
	// PeelUnknown means no peel information is available;
	// consumers that need the peeled target
	// must consult object storage.
	PeelUnknown PeelState = iota

	// PeelNone means the referent is known not to be peelable,
	// such as a reference pointing directly at a commit.
	PeelNone

	// PeelTo means the referent peels to PeeledID.
	PeelTo
)

// Direct points directly to an object ID.
//
// Labels: MT-Unsafe.
type Direct struct {
	RefName string
	ID      objectid.ObjectID

	// PeelState describes the cached peel knowledge for the referent.
	PeelState PeelState

	// PeeledID is the fully peeled target.
	// It is meaningful only when PeelState is [PeelTo].
	PeeledID objectid.ObjectID
}

// Name returns the fully-qualified reference name.
func (ref Direct) Name() string {
	return ref.RefName
}

func (Direct) isRef() {}
