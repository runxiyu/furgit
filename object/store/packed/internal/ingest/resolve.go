package ingest

import (
	"fmt"
	"io"
	"sync"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packfile/delta"
	"lindenii.org/go/furgit/internal/progress"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// adjacency maps each resolvable base to its delta children:
// ofs-deltas keyed by base offset, ref-deltas keyed by base object ID.
type adjacency struct {
	byOffset map[int][]int
	byOID    map[id.ObjectID][]int
}

// resolveDeltas resolves every delta record into a final object ID and type,
// completing thin packs from the external base reader when required.
func (ingestion *ingestion) resolveDeltas() error {
	meter := progress.New(progress.Options{
		Writer:     ingestion.opts.Progress,
		Title:      "resolving deltas",
		Total:      ingestion.countDeltas(),
		Delay:      0,
		Sparse:     false,
		Throughput: false,
	})

	adjacency := ingestion.buildAdjacency()

	err := ingestion.resolveFrom(ingestion.resolvedRoots(), adjacency, meter)
	if err != nil {
		return err
	}

	external := ingestion.unresolvedExternalBases()

	switch {
	case len(external) == 0 && ingestion.countUnresolved() > 0:
		return fmt.Errorf("%w: unresolvable delta entries", ErrMalformedPack)
	case len(external) > 0:
		err = ingestion.fixThin(external, adjacency, meter)
		if err != nil {
			return err
		}
	}

	meter.Stop("done")

	return nil
}

// buildAdjacency indexes every delta record by its base,
// so a resolved base can find the children that delta against it.
func (ingestion *ingestion) buildAdjacency() adjacency {
	out := adjacency{
		byOffset: make(map[int][]int),
		byOID:    make(map[id.ObjectID][]int),
	}

	for index := range ingestion.records {
		rec := &ingestion.records[index]

		switch rec.packedType {
		case packfile.EntryTypeOfsDelta:
			out.byOffset[rec.baseOffset] = append(out.byOffset[rec.baseOffset], index)
		case packfile.EntryTypeRefDelta:
			out.byOID[rec.baseOID] = append(out.byOID[rec.baseOID], index)
		case packfile.EntryTypeInvalid,
			packfile.EntryTypeCommit,
			packfile.EntryTypeTree,
			packfile.EntryTypeBlob,
			packfile.EntryTypeTag,
			packfile.EntryTypeFuture:
		}
	}

	return out
}

// item is a delta record awaiting resolution, with its delta-chain depth.
type item struct {
	index int
	depth int
}

// resolver resolves deltas concurrently over a shared LIFO work stack.
//
// Each item is one delta child;
// a worker materializes its base from the cache, re-deriving on a miss,
// resolves the child,
// and pushes the child's own delta children.
// Workers park while the stack is empty but others are still working,
// and exit once it is empty and none are.
type resolver struct {
	ingestion *ingestion
	adjacency adjacency
	meter     *progress.Meter

	mu       sync.Mutex
	cond     *sync.Cond
	stack    []item
	active   int
	firstErr error
}

func (ingestion *ingestion) resolveFrom(roots []int, adjacency adjacency, meter *progress.Meter) error {
	var seed []item

	for _, root := range roots {
		rec := &ingestion.records[root]
		for _, group := range [2][]int{adjacency.byOffset[rec.offset], adjacency.byOID[rec.oid]} {
			for _, child := range group {
				seed = append(seed, item{index: child, depth: 1})
			}
		}
	}

	if len(seed) == 0 {
		return nil
	}

	res := &resolver{
		ingestion: ingestion,
		adjacency: adjacency,
		meter:     meter,
		stack:     seed,
	}
	res.cond = sync.NewCond(&res.mu)

	return res.run(ingestion.workers)
}

func (res *resolver) run(workers int) error {
	if workers <= 1 {
		res.worker()

		return res.firstErr
	}

	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			res.worker()
		})
	}

	wg.Wait()

	return res.firstErr
}

func (res *resolver) worker() {
	for {
		res.mu.Lock()

		for len(res.stack) == 0 && res.active > 0 && res.firstErr == nil {
			res.cond.Wait()
		}

		if res.firstErr != nil || len(res.stack) == 0 {
			res.mu.Unlock()

			return
		}

		it := res.stack[len(res.stack)-1]
		res.stack = res.stack[:len(res.stack)-1]
		res.active++
		res.mu.Unlock()

		kids, err := res.process(it)

		res.mu.Lock()
		res.active--

		if err != nil && res.firstErr == nil {
			res.firstErr = err
		}

		if res.firstErr == nil {
			res.stack = append(res.stack, kids...)
		}

		if res.firstErr != nil || len(kids) > 0 || (res.active == 0 && len(res.stack) == 0) {
			res.cond.Broadcast()
		}

		res.mu.Unlock()
	}
}

// process resolves one delta child and returns its own delta children.
func (res *resolver) process(it item) ([]item, error) {
	err := res.ingestion.ctx.Err()
	if err != nil {
		return nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	rec := &res.ingestion.records[it.index]

	parent, ok := res.ingestion.baseRecordIndex(rec)
	if !ok {
		return nil, fmt.Errorf("%w: entry at %d: base unavailable while resolving", ErrMalformedPack, rec.offset)
	}

	baseType, baseContent, err := res.ingestion.materialize(parent)
	if err != nil {
		return nil, err
	}

	err = res.ingestion.resolveOneChild(it.index, baseType, baseContent, res.meter)
	if err != nil {
		return nil, err
	}

	return res.childItems(it.index, it.depth+1)
}

// childItems returns the delta children of a just-resolved record at depth.
func (res *resolver) childItems(index, depth int) ([]item, error) {
	rec := &res.ingestion.records[index]

	var kids []item

	for _, group := range [2][]int{res.adjacency.byOffset[rec.offset], res.adjacency.byOID[rec.oid]} {
		for _, child := range group {
			if depth > delta.MaxChainDepth {
				return nil, fmt.Errorf("%w: entry at %d: delta chain too deep", ErrMalformedPack, res.ingestion.records[child].offset)
			}

			kids = append(kids, item{index: child, depth: depth})
		}
	}

	return kids, nil
}

func (ingestion *ingestion) resolveOneChild(index int, baseType typ.Type, baseContent []byte, meter *progress.Meter) error {
	rec := &ingestion.records[index]

	content, err := ingestion.applyDelta(index, baseContent)
	if err != nil {
		return err
	}

	oid, err := ingestion.hashObject(baseType, content)
	if err != nil {
		return err
	}

	rec.oid = oid
	rec.resolved = true

	ingestion.byOID.Store(oid, index)
	ingestion.baseCache.Add(baseCacheKey{offset: rec.offset}, cachedContent{objectType: baseType, content: content})

	ingestion.deltasResolved.Add(1)
	meter.Add(1, 0)

	return nil
}

// materialize returns the inflated content of an already-resolved record,
// from the base cache,
// or re-derived from the nearest cached or base ancestor on a miss.
func (ingestion *ingestion) materialize(index int) (typ.Type, []byte, error) {
	var (
		zero     typ.Type
		chain    []int
		base     []byte
		baseType typ.Type
	)

	cur := index

	for {
		rec := &ingestion.records[cur]

		if cached, ok := ingestion.baseCache.Get(baseCacheKey{offset: rec.offset}); ok {
			base = cached.content
			baseType = cached.objectType

			break
		}

		if rec.packedType.IsBase() {
			objectType, err := rec.packedType.ObjectType()
			if err != nil {
				return zero, nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
			}

			content, err := ingestion.inflateRecord(cur)
			if err != nil {
				return zero, nil, err
			}

			base = content
			baseType = objectType

			break
		}

		if len(chain) >= delta.MaxChainDepth {
			return zero, nil, fmt.Errorf("%w: entry at %d: delta chain too deep", ErrMalformedPack, rec.offset)
		}

		chain = append(chain, cur)

		next, ok := ingestion.baseRecordIndex(rec)
		if !ok {
			return zero, nil, fmt.Errorf("%w: entry at %d: base unavailable while reconstructing", ErrMalformedPack, rec.offset)
		}

		cur = next
	}

	for i := len(chain) - 1; i >= 0; i-- {
		content, err := ingestion.applyDelta(chain[i], base)
		if err != nil {
			return zero, nil, err
		}

		ingestion.baseCache.Add(baseCacheKey{offset: ingestion.records[chain[i]].offset}, cachedContent{objectType: baseType, content: content})

		base = content
	}

	return baseType, base, nil
}

func (ingestion *ingestion) applyDelta(index int, baseContent []byte) ([]byte, error) {
	rec := &ingestion.records[index]

	deltaPayload, err := ingestion.inflateRecord(index)
	if err != nil {
		return nil, err
	}

	baseSize, resultSize, _, err := delta.ParseHeaderSizes(deltaPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	limit := ingestion.opts.MaxObjectSize
	if limit > 0 && resultSize > uint64(limit) {
		return nil, fmt.Errorf("%w: entry at %d: result size %d exceeds limit %d", store.ErrObjectTooLarge, rec.offset, resultSize, limit)
	}

	if baseSize != uint64(len(baseContent)) {
		return nil, fmt.Errorf("%w: entry at %d: delta base size mismatch", ErrMalformedPack, rec.offset)
	}

	content, err := delta.Apply(baseContent, deltaPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	if uint64(len(content)) != resultSize {
		return nil, fmt.Errorf("%w: entry at %d: delta result size mismatch", ErrMalformedPack, rec.offset)
	}

	return content, nil
}

func (ingestion *ingestion) baseRecordIndex(rec *record) (int, bool) {
	switch rec.packedType {
	case packfile.EntryTypeOfsDelta:
		index, ok := ingestion.byOffset[rec.baseOffset]

		return index, ok
	case packfile.EntryTypeRefDelta:
		index, ok := ingestion.byOID.Load(rec.baseOID)

		return index, ok
	case packfile.EntryTypeInvalid,
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
		packfile.EntryTypeFuture:
	}

	return 0, false
}

// inflateRecord inflates one record's payload from the temporary pack file.
func (ingestion *ingestion) inflateRecord(index int) ([]byte, error) {
	rec := &ingestion.records[index]

	offset := int64(rec.dataOffset())
	compressedLen := int64(rec.packedLen - rec.headerLen)
	size := rec.declaredSize

	zr, err := zlib.NewReader(io.NewSectionReader(ingestion.packFile, offset, compressedLen))
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	defer func() { _ = zr.Close() }()

	out := make([]byte, size)

	_, err = io.ReadFull(zr, out)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	return out, nil
}

// hashObject computes the object ID of one resolved object.
func (ingestion *ingestion) hashObject(objectType typ.Type, content []byte) (id.ObjectID, error) {
	var zero id.ObjectID

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(header.Append(nil, objectType, len(content)))
	_, _ = hashImpl.Write(content)

	oid, err := ingestion.objectFormat.FromBytes(hashImpl.Sum(nil))
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	return oid, nil
}

// resolvedRoots returns the indices of every currently resolved record.
func (ingestion *ingestion) resolvedRoots() []int {
	var roots []int

	for index := range ingestion.records {
		if ingestion.records[index].resolved {
			roots = append(roots, index)
		}
	}

	return roots
}

// countDeltas returns the number of delta records.
func (ingestion *ingestion) countDeltas() int {
	return ingestion.deltaCount
}

// countUnresolved returns the number of records that remain unresolved.
//
// Every base is resolved during scanning or thin completion,
// so the unresolved records are exactly the unresolved deltas:
// the delta records minus those already resolved.
func (ingestion *ingestion) countUnresolved() int {
	return ingestion.deltaCount - int(ingestion.deltasResolved.Load())
}

// unresolvedExternalBases returns the unique base object IDs
// of unresolved ref-deltas whose base is not present in the pack,
// in first-reference order.
func (ingestion *ingestion) unresolvedExternalBases() []id.ObjectID {
	seen := make(map[id.ObjectID]struct{})

	out := make([]id.ObjectID, 0, ingestion.deltaCount-int(ingestion.deltasResolved.Load()))

	for index := range ingestion.records {
		rec := &ingestion.records[index]
		if rec.resolved || rec.packedType != packfile.EntryTypeRefDelta {
			continue
		}

		if _, ok := ingestion.byOID.Load(rec.baseOID); ok {
			continue
		}

		if _, ok := seen[rec.baseOID]; ok {
			continue
		}

		seen[rec.baseOID] = struct{}{}
		out = append(out, rec.baseOID)
	}

	return out
}
