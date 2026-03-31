package reachability

import "fmt"

// Domain specifies which graph edges are traversed.
type Domain uint8

const (
	// DomainCommits traverses commit-parent edges and annotated-tag target edges.
	DomainCommits Domain = iota
	// DomainObjects traverses full commit/tree/blob objects.
	DomainObjects
)

func validateDomain(domain Domain) error {
	switch domain {
	case DomainCommits, DomainObjects:
		return nil
	default:
		return fmt.Errorf("reachability: invalid domain %d", domain)
	}
}
