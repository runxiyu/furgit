package furgit

import (
	"fmt"
	"iter"
)

// ReachabilityMode controls which object types are walked.
type ReachabilityMode uint8

const (
	// ReachabilityCommitsOnly walks only commit objects.
	ReachabilityCommitsOnly ReachabilityMode = iota
	// ReachabilityAllObjects walks commits, trees, blobs, and tags reachable
	// from the commits in Wants.
	ReachabilityAllObjects
)

// ReachabilityQuery describes a want/have reachability walk.
//
// ReachableObjects returns objects reachable from Wants. If Mode is
// ReachabilityCommitsOnly, non-commit Wants are ignored except for tags,
// which are peeled to their target.
type ReachabilityQuery struct {
	Wants []Hash
	Haves []Hash
	Mode  ReachabilityMode

	// StopAtHaves prunes traversal when an object is reachable from Haves.
	StopAtHaves bool
}

// ReachableObject reports a reachable object and whether it is also reachable
// from the Have set.
type ReachableObject struct {
	ID     Hash
	Type   ObjectType
	InHave bool
}

// ReachabilityWalk is a single-use reachability iterator.
// After iterating, Err reports any error encountered during the walk.
type ReachabilityWalk struct {
	repo  *Repository
	query ReachabilityQuery
	err   error

	haveInit bool
	haveErr  error
	haveSet  map[Hash]struct{}
}

// ReachableObjects returns a single-use iterator over objects reachable from
// query.Wants.
//
// It yields ReachableObject values; InHave is true when the object is also
// reachable from query.Haves.
func (repo *Repository) ReachableObjects(query ReachabilityQuery) (*ReachabilityWalk, error) {
	if repo == nil {
		return nil, ErrInvalidObject
	}
	switch query.Mode {
	case ReachabilityCommitsOnly, ReachabilityAllObjects:
	default:
		return nil, ErrInvalidObject
	}
	for _, id := range query.Wants {
		if id.algo != repo.hashAlgo {
			return nil, fmt.Errorf("furgit: reachability: want hash algorithm mismatch")
		}
	}
	for _, id := range query.Haves {
		if id.algo != repo.hashAlgo {
			return nil, fmt.Errorf("furgit: reachability: have hash algorithm mismatch")
		}
	}
	return &ReachabilityWalk{
		repo:  repo,
		query: query,
	}, nil
}

// Seq returns the iterator.
func (w *ReachabilityWalk) Seq() iter.Seq[ReachableObject] {
	return func(yield func(ReachableObject) bool) {
		if w == nil || w.repo == nil {
			w.err = ErrInvalidObject
			return
		}
		haveSet, err := w.ensureHaveSet()
		if err != nil {
			w.err = err
			return
		}

		wantWalk := reachabilityWalker{
			repo:        w.repo,
			mode:        w.query.Mode,
			seenCommits: make(map[Hash]struct{}),
			seenObjects: make(map[Hash]struct{}),
			haveSet:     haveSet,
			stopAtHaves: w.query.StopAtHaves,
		}
		if err := wantWalk.walkRoots(w.query.Wants, func(obj ReachableObject) bool {
			return yield(obj)
		}); err != nil {
			w.err = err
			return
		}
	}
}

// Err reports the first error encountered by the iterator.
func (w *ReachabilityWalk) Err() error {
	if w == nil {
		return ErrInvalidObject
	}
	return w.err
}

// HaveContains reports whether id is reachable from Haves.
func (w *ReachabilityWalk) HaveContains(id Hash) (bool, error) {
	if w == nil || w.repo == nil {
		return false, ErrInvalidObject
	}
	haveSet, err := w.ensureHaveSet()
	if err != nil {
		return false, err
	}
	_, ok := haveSet[id]
	return ok, nil
}

func (w *ReachabilityWalk) ensureHaveSet() (map[Hash]struct{}, error) {
	if w.haveInit {
		return w.haveSet, w.haveErr
	}
	w.haveInit = true
	w.haveSet = make(map[Hash]struct{})
	if len(w.query.Haves) == 0 {
		return w.haveSet, nil
	}
	haveWalk := reachabilityWalker{
		repo:           w.repo,
		mode:           w.query.Mode,
		seenCommits:    make(map[Hash]struct{}),
		seenObjects:    make(map[Hash]struct{}),
		recordHaveOnly: true,
		haveSet:        w.haveSet,
	}
	if err := haveWalk.walkRoots(w.query.Haves, nil); err != nil {
		w.haveErr = err
		return nil, err
	}
	return w.haveSet, nil
}

type reachabilityWalker struct {
	repo *Repository
	mode ReachabilityMode

	seenCommits map[Hash]struct{}
	seenObjects map[Hash]struct{}

	haveSet        map[Hash]struct{}
	recordHaveOnly bool
	stopAtHaves    bool

	cg     *commitGraph
	cgInit bool
}

func (rw *reachabilityWalker) initCommitGraph() {
	if rw.cgInit {
		return
	}
	rw.cgInit = true
	cg, err := rw.repo.CommitGraph()
	if err == nil {
		rw.cg = cg
	}
}

func (rw *reachabilityWalker) walkRoots(roots []Hash, emit func(ReachableObject) bool) error {
	for _, id := range roots {
		if err := rw.walkObject(id, emit); err != nil {
			return err
		}
	}
	return nil
}

func (rw *reachabilityWalker) walkObject(id Hash, emit func(ReachableObject) bool) error {
	if rw.stopAtHaves {
		if _, ok := rw.haveSet[id]; ok {
			return nil
		}
	}
	if rw.recordHaveOnly {
		if _, ok := rw.haveSet[id]; ok {
			return nil
		}
	} else {
		if _, ok := rw.seenObjects[id]; ok {
			return nil
		}
	}

	rw.initCommitGraph()
	if rw.cg != nil {
		if pos, ok := rw.cg.CommitPosition(id); ok {
			return rw.walkCommitByPos(pos, id, emit)
		}
	}

	ty, body, err := rw.repo.ReadObjectTypeRaw(id)
	if err != nil {
		return err
	}

	switch ty {
	case ObjectTypeCommit:
		return rw.walkCommitBody(id, body, emit)
	case ObjectTypeTree:
		if rw.mode != ReachabilityAllObjects {
			return nil
		}
		return rw.walkTreeBody(id, body, emit)
	case ObjectTypeBlob:
		if rw.mode != ReachabilityAllObjects {
			return nil
		}
		return rw.emitObject(id, ObjectTypeBlob, emit)
	case ObjectTypeTag:
		return rw.walkTagBody(id, body, emit)
	default:
		return ErrInvalidObject
	}
}

func (rw *reachabilityWalker) walkCommitByPos(pos uint32, id Hash, emit func(ReachableObject) bool) error {
	if _, ok := rw.seenCommits[id]; ok {
		return nil
	}
	rw.seenCommits[id] = struct{}{}

	cc, err := rw.cg.CommitAt(pos)
	if err != nil {
		return err
	}

	if err := rw.emitObject(id, ObjectTypeCommit, emit); err != nil {
		return err
	}

	if rw.mode == ReachabilityAllObjects {
		if err := rw.walkTreeByID(cc.Tree, emit); err != nil {
			return err
		}
	}

	for _, parentPos := range cc.Parents {
		parentID, err := rw.cg.OIDAt(parentPos)
		if err != nil {
			return err
		}
		if err := rw.walkObject(parentID, emit); err != nil {
			return err
		}
	}
	return nil
}

func (rw *reachabilityWalker) walkCommitBody(id Hash, body []byte, emit func(ReachableObject) bool) error {
	if _, ok := rw.seenCommits[id]; ok {
		return nil
	}
	rw.seenCommits[id] = struct{}{}

	commit, err := parseCommit(id, body, rw.repo)
	if err != nil {
		return err
	}
	if err := rw.emitObject(id, ObjectTypeCommit, emit); err != nil {
		return err
	}
	if rw.mode == ReachabilityAllObjects {
		if err := rw.walkTreeByID(commit.Tree, emit); err != nil {
			return err
		}
	}
	for _, parent := range commit.Parents {
		if err := rw.walkObject(parent, emit); err != nil {
			return err
		}
	}
	return nil
}

func (rw *reachabilityWalker) walkTagBody(id Hash, body []byte, emit func(ReachableObject) bool) error {
	tag, err := parseTag(id, body, rw.repo)
	if err != nil {
		return err
	}
	if rw.mode == ReachabilityAllObjects {
		if err := rw.emitObject(id, ObjectTypeTag, emit); err != nil {
			return err
		}
	}
	if tag.TargetType == ObjectTypeCommit {
		return rw.walkObject(tag.Target, emit)
	}
	if rw.mode == ReachabilityAllObjects {
		return rw.walkObject(tag.Target, emit)
	}
	return nil
}

func (rw *reachabilityWalker) walkTreeByID(id Hash, emit func(ReachableObject) bool) error {
	if rw.mode != ReachabilityAllObjects {
		return nil
	}
	if _, ok := rw.seenObjects[id]; ok && !rw.recordHaveOnly {
		return nil
	}
	ty, body, err := rw.repo.ReadObjectTypeRaw(id)
	if err != nil {
		return err
	}
	if ty != ObjectTypeTree {
		return ErrInvalidObject
	}
	return rw.walkTreeBody(id, body, emit)
}

func (rw *reachabilityWalker) walkTreeBody(id Hash, body []byte, emit func(ReachableObject) bool) error {
	if rw.mode != ReachabilityAllObjects {
		return nil
	}
	tree, err := parseTree(id, body, rw.repo)
	if err != nil {
		return err
	}
	if err := rw.emitObject(id, ObjectTypeTree, emit); err != nil {
		return err
	}
	for _, entry := range tree.Entries {
		switch entry.Mode {
		case FileModeDir:
			if err := rw.walkTreeByID(entry.ID, emit); err != nil {
				return err
			}
		case FileModeGitlink:
			// IIRC Gitlinks are references to external repositories
			// and do not imply reachability of the target commit...
			continue
		default:
			if err := rw.emitObject(entry.ID, ObjectTypeBlob, emit); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rw *reachabilityWalker) emitObject(id Hash, ty ObjectType, emit func(ReachableObject) bool) error {
	if rw.recordHaveOnly {
		rw.haveSet[id] = struct{}{}
		return nil
	}
	if _, ok := rw.seenObjects[id]; ok {
		return nil
	}
	rw.seenObjects[id] = struct{}{}
	inHave := false
	if _, ok := rw.haveSet[id]; ok {
		inHave = true
	}
	if emit != nil {
		if !emit(ReachableObject{ID: id, Type: ty, InHave: inHave}) {
			return nil
		}
	}
	return nil
}
