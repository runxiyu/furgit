package reachability

import (
	"errors"
	"fmt"
	"iter"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	"codeberg.org/lindenii/furgit/objecttype"
)

// Domain specifies which graph edges are traversed.
type Domain uint8

const (
	// DomainCommits traverses commit-parent edges and annotated-tag target edges.
	DomainCommits Domain = iota
	// DomainObjects traverses full commit/tree/blob objects.
	DomainObjects
)

// Reachability provides graph traversal over objects in one object store.
//
// It is not safe for concurrent use.
type Reachability struct {
	Store objectstore.Store
}

// New builds a Reachability  over one object store.
func New(store objectstore.Store) *Reachability {
	return &Reachability{Store: store}
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

	walk := r.Walk(DomainCommits, nil, map[objectid.ObjectID]struct{}{descendantCommit: {}})
	for id := range walk.Seq() {
		if id == ancestorCommit {
			return true, nil
		}
	}
	if err := walk.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// CheckConnected verifies that all objects reachable from wants (under the
// selected domain) can be fully traversed without missing-object/type/parse
// errors, excluding subgraphs rooted at haves.
func (r *Reachability) CheckConnected(domain Domain, haves, wants map[objectid.ObjectID]struct{}) error {
	walk := r.Walk(domain, haves, wants)
	for range walk.Seq() {
	}
	return walk.Err()
}

// Walk creates one single-use traversal over the selected domain.
func (r *Reachability) Walk(domain Domain, haves, wants map[objectid.ObjectID]struct{}) *Walk {
	walk := &Walk{
		reachability: r,
		domain:       domain,
		haves:        haves,
		wants:        wants,
	}
	if err := validateDomain(domain); err != nil {
		walk.err = err
	}
	return walk
}

// ErrObjectMissing indicates that a referenced object is absent from the store.
type ErrObjectMissing struct {
	OID objectid.ObjectID
}

func (e *ErrObjectMissing) Error() string {
	return fmt.Sprintf("reachability: missing object %s", e.OID)
}

// ErrObjectType indicates that a referenced object has a different type than
// what traversal expected on that edge.
type ErrObjectType struct {
	OID  objectid.ObjectID
	Got  objecttype.Type
	Want objecttype.Type
}

func (e *ErrObjectType) Error() string {
	gotName, gotOK := objecttype.Name(e.Got)
	if !gotOK {
		gotName = fmt.Sprintf("type(%d)", e.Got)
	}
	wantName, wantOK := objecttype.Name(e.Want)
	if !wantOK {
		wantName = fmt.Sprintf("type(%d)", e.Want)
	}
	return fmt.Sprintf("reachability: object %s has type %s, want %s", e.OID, gotName, wantName)
}

// Walk is one single-use iterator-style traversal.
type Walk struct {
	reachability *Reachability
	domain       Domain
	haves        map[objectid.ObjectID]struct{}
	wants        map[objectid.ObjectID]struct{}

	seqUsed bool
	err     error
}

// Seq returns the traversal sequence. It is single-use.
func (walk *Walk) Seq() iter.Seq[objectid.ObjectID] {
	if walk.seqUsed {
		return func(yield func(objectid.ObjectID) bool) {
			_ = yield
			if walk.err == nil {
				walk.err = errors.New("reachability: walk sequence already consumed")
			}
		}
	}
	walk.seqUsed = true
	return func(yield func(objectid.ObjectID) bool) {
		if walk.err != nil {
			return
		}
		stack := walk.initialStack()
		var err error
		visited := make(map[objectid.ObjectID]struct{}, len(stack))
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if containsOID(walk.haves, item.id) {
				continue
			}
			if _, ok := visited[item.id]; ok {
				continue
			}
			visited[item.id] = struct{}{}

			var next []walkItem
			next, err = walk.expand(item)
			if err != nil {
				walk.err = err
				return
			}
			if !yield(item.id) {
				return
			}
			stack = append(stack, next...)
		}
	}
}

// Err returns the terminal error, if any, once Seq has been consumed.
func (walk *Walk) Err() error {
	return walk.err
}

type walkItem struct {
	id   objectid.ObjectID
	want objecttype.Type
}

func (walk *Walk) initialStack() []walkItem {
	if len(walk.wants) == 0 {
		return nil
	}
	stack := make([]walkItem, 0, len(walk.wants))
	for want := range walk.wants {
		stack = append(stack, walkItem{id: want, want: objecttype.TypeInvalid})
	}
	return stack
}

func (walk *Walk) expand(item walkItem) ([]walkItem, error) {
	if walk.domain == DomainCommits {
		return walk.expandCommits(item)
	}
	return walk.expandObjects(item)
}

func (walk *Walk) expandCommits(item walkItem) ([]walkItem, error) {
	ty, err := walk.readHeaderType(item.id)
	if err != nil {
		return nil, err
	}
	switch ty {
	case objecttype.TypeCommit:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}
		commit, err := object.ParseCommit(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}
		next := make([]walkItem, 0, len(commit.Parents))
		for _, parent := range commit.Parents {
			next = append(next, walkItem{id: parent, want: objecttype.TypeInvalid})
		}
		return next, nil
	case objecttype.TypeTag:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}
		tag, err := object.ParseTag(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}
		return []walkItem{{id: tag.Target, want: objecttype.TypeInvalid}}, nil
	case objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeInvalid,
		objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: objecttype.TypeCommit}
	}
	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}

func (walk *Walk) expandObjects(item walkItem) ([]walkItem, error) {
	ty, err := walk.readHeaderType(item.id)
	if err != nil {
		return nil, err
	}
	if item.want != objecttype.TypeInvalid && ty != item.want {
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: item.want}
	}

	switch ty {
	case objecttype.TypeBlob:
		return nil, nil
	case objecttype.TypeCommit:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}
		commit, err := object.ParseCommit(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}
		next := make([]walkItem, 0, len(commit.Parents)+1)
		next = append(next, walkItem{id: commit.Tree, want: objecttype.TypeTree})
		for _, parent := range commit.Parents {
			next = append(next, walkItem{id: parent, want: objecttype.TypeCommit})
		}
		return next, nil
	case objecttype.TypeTree:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}
		tree, err := object.ParseTree(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}
		next := make([]walkItem, 0, len(tree.Entries))
		for _, entry := range tree.Entries {
			switch entry.Mode {
			case object.FileModeGitlink:
				continue
			case object.FileModeDir:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeTree})
			case object.FileModeRegular, object.FileModeExecutable, object.FileModeSymlink:
				next = append(next, walkItem{id: entry.ID, want: objecttype.TypeBlob})
			}
		}
		return next, nil
	case objecttype.TypeTag:
		content, err := walk.readBytesContent(item.id)
		if err != nil {
			return nil, err
		}
		tag, err := object.ParseTag(content, item.id.Algorithm())
		if err != nil {
			return nil, err
		}
		return []walkItem{{id: tag.Target, want: tag.TargetType}}, nil
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return nil, &ErrObjectType{OID: item.id, Got: ty, Want: item.want}
	}
	return nil, fmt.Errorf("reachability: unreachable object type %d", ty)
}

func (r *Reachability) peelRootToDomain(id objectid.ObjectID, domain Domain) (objectid.ObjectID, error) {
	if err := validateDomain(domain); err != nil {
		return objectid.ObjectID{}, err
	}
	for {
		ty, err := r.readHeaderType(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}
		if ty != objecttype.TypeTag {
			if domain == DomainCommits && ty != objecttype.TypeCommit {
				return objectid.ObjectID{}, &ErrObjectType{OID: id, Got: ty, Want: objecttype.TypeCommit}
			}
			return id, nil
		}

		content, err := r.readBytesContent(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}
		tag, err := object.ParseTag(content, id.Algorithm())
		if err != nil {
			return objectid.ObjectID{}, err
		}
		id = tag.Target
	}
}

func validateDomain(domain Domain) error {
	switch domain {
	case DomainCommits, DomainObjects:
		return nil
	default:
		return fmt.Errorf("reachability: invalid domain %d", domain)
	}
}

func containsOID(set map[objectid.ObjectID]struct{}, id objectid.ObjectID) bool {
	if len(set) == 0 {
		return false
	}
	_, ok := set[id]
	return ok
}

// The following helpers exist because we don't have unified error handling across the entire project.
// This will be fixed later.

func (walk *Walk) readHeaderType(id objectid.ObjectID) (objecttype.Type, error) {
	return walk.reachability.readHeaderType(id)
}

func (r *Reachability) readHeaderType(id objectid.ObjectID) (objecttype.Type, error) {
	ty, _, err := r.Store.ReadHeader(id)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return objecttype.TypeInvalid, &ErrObjectMissing{OID: id}
		}
		return objecttype.TypeInvalid, err
	}
	return ty, nil
}

func (walk *Walk) readBytesContent(id objectid.ObjectID) ([]byte, error) {
	content, err := walk.reachability.readBytesContent(id)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (r *Reachability) readBytesContent(id objectid.ObjectID) ([]byte, error) {
	_, content, err := r.Store.ReadBytesContent(id)
	if err != nil {
		if errors.Is(err, objectstore.ErrObjectNotFound) {
			return nil, &ErrObjectMissing{OID: id}
		}
		return nil, err
	}
	return content, nil
}
