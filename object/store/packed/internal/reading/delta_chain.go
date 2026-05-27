package reading

import objecttype "lindenii.org/go/furgit/object/type"

// deltaChain describes how to reconstruct one requested object.
type deltaChain struct {
	// baseLoc points to the innermost base object.
	baseLoc location
	// baseType is the canonical object type resolved from baseLoc.
	baseType objecttype.Type
	// deltas contains delta objects from target down toward base.
	deltas []deltaNode
}
