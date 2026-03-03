package reachability

// Domain specifies which graph edges are traversed.
type Domain uint8

const (
	// DomainCommits traverses commit-parent edges and annotated-tag target edges.
	DomainCommits Domain = iota
	// DomainObjects traverses full commit/tree/blob objects.
	DomainObjects
)
