package packed

import "codeberg.org/lindenii/furgit/objecttype"

// deltaChain describes how to reconstruct one requested object.
type deltaChain struct {
	// baseLoc points to the innermost base object.
	baseLoc location
	// baseType is the canonical object type resolved from baseLoc.
	baseType objecttype.Type
	// deltas contains delta objects from target down toward base.
	deltas []deltaNode
}
