// Package reachability traverses the object graph to test relationships and emit object lists.
package reachability

import (
	"errors"

	"codeberg.org/lindenii/furgit/format/commitgraph"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Reachability provides graph traversal over objects in one object store.
//
// It is not safe for concurrent use.
type Reachability struct {
	store objectstore.Store
	graph *commitgraph.Reader
}

// New builds a Reachability  over one object store.
func New(store objectstore.Store) *Reachability {
	return &Reachability{store: store}
}

// NewWithCommitGraph builds a Reachability over one object store with an
// optional commit-graph reader for faster commit-domain traversal.
func NewWithCommitGraph(store objectstore.Store, graph *commitgraph.Reader) *Reachability {
	return &Reachability{store: store, graph: graph}
}

// IsAncestor reports whether ancestor is reachable from descendant via commit
// parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (r *Reachability) IsAncestor(ancestor, descendant objectid.ObjectID) (bool, error) {
	ancestorCommit, err := r.peelRootToDomain(ancestor, DomainCommits)
	if err != nil {
		return false, err
	}

	descendantCommit, err := r.peelRootToDomain(descendant, DomainCommits)
	if err != nil {
		return false, err
	}

	if ancestorCommit == descendantCommit {
		return true, nil
	}

	graphResult, graphUsed, err := r.isAncestorGraph(ancestorCommit, descendantCommit)
	if err != nil {
		return false, err
	}

	if graphUsed {
		return graphResult, nil
	}

	walk := r.Walk(DomainCommits, nil, map[objectid.ObjectID]struct{}{descendantCommit: {}})
	for id := range walk.Seq() {
		if id == ancestorCommit {
			return true, nil
		}
	}

	err = walk.Err()
	if err != nil {
		return false, err
	}

	return false, nil
}

// CheckConnected verifies that all objects reachable from wants (under the
// selected domain) can be fully traversed without missing-object/type/parse
// errors, excluding subgraphs rooted at haves.
//
// Even with commit-graph acceleration available, each visited commit is
// still validated against the object store.
func (r *Reachability) CheckConnected(domain Domain, haves, wants map[objectid.ObjectID]struct{}) error {
	walk := r.Walk(domain, haves, wants)

	walk.strict = true
	for range walk.Seq() {
	}

	return walk.Err()
}

// Walk creates one single-use traversal over the selected domain.
//
// In DomainCommits, when a commit-graph reader is attached, parent expansion
// may use commit-graph metadata for speed.
func (r *Reachability) Walk(domain Domain, haves, wants map[objectid.ObjectID]struct{}) *Walk {
	walk := &Walk{
		reachability: r,
		domain:       domain,
		haves:        haves,
		wants:        wants,
	}

	err := validateDomain(domain)
	if err != nil {
		walk.err = err
	}

	return walk
}

func (r *Reachability) isAncestorGraph(ancestor, descendant objectid.ObjectID) (bool, bool, error) {
	if r.graph == nil {
		return false, false, nil
	}

	ancestorPos, err := r.graph.Lookup(ancestor)
	if err != nil {
		var notFound *commitgraph.ErrNotFound
		if errors.As(err, &notFound) {
			return false, false, nil
		}

		return false, true, err
	}

	descendantPos, err := r.graph.Lookup(descendant)
	if err != nil {
		var notFound *commitgraph.ErrNotFound
		if errors.As(err, &notFound) {
			return false, false, nil
		}

		return false, true, err
	}

	ancestorCommit, err := r.graph.CommitAt(ancestorPos)
	if err != nil {
		return false, true, err
	}

	ancestorGeneration := ancestorCommit.GenerationV2
	stack := []commitgraph.Position{descendantPos}
	visited := make(map[commitgraph.Position]struct{}, 64)

	for len(stack) > 0 {
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := visited[pos]; ok {
			continue
		}

		visited[pos] = struct{}{}

		if pos == ancestorPos {
			return true, true, nil
		}

		commit, err := r.graph.CommitAt(pos)
		if err != nil {
			return false, true, err
		}

		if commit.GenerationV2 < ancestorGeneration {
			continue
		}

		if commit.Parent1.Valid {
			stack = append(stack, commit.Parent1.Pos)
		}

		if commit.Parent2.Valid {
			stack = append(stack, commit.Parent2.Pos)
		}

		stack = append(stack, commit.ExtraParents...)
	}

	return false, true, nil
}
