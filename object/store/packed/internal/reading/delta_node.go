package reading

// deltaNode describes one delta object in a reconstruction chain.
type deltaNode struct {
	// loc identifies the delta object's pack location.
	loc location
	// dataOffset points to the start of the delta zlib payload in pack.
	dataOffset int
}
