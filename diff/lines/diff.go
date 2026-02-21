// Package lines provides routines to perform line-based diffs.
package lines

import "bytes"

// Diff performs a line-based diff.
// Lines are bytes up to and including '\n' (final line may lack '\n').
func Diff(oldB, newB []byte) ([]Chunk, error) {
	type lineRef struct {
		base  []byte
		start int
		end   int
	}

	split := func(b []byte) []lineRef {
		if len(b) == 0 {
			return nil
		}
		var res []lineRef
		start := 0
		for i := range b {
			if b[i] == '\n' {
				res = append(res, lineRef{base: b, start: start, end: i + 1})
				start = i + 1
			}
		}
		if start < len(b) {
			res = append(res, lineRef{base: b, start: start, end: len(b)})
		}
		return res
	}

	oldLines := split(oldB)
	newLines := split(newB)

	n := len(oldLines)
	m := len(newLines)
	if n == 0 && m == 0 {
		return nil, nil
	}

	idOf := make(map[string]int)
	nextID := 0
	oldIDs := make([]int, n)
	for i, ln := range oldLines {
		key := string(ln.base[ln.start:ln.end])
		id, ok := idOf[key]
		if !ok {
			id = nextID
			idOf[key] = id
			nextID++
		}
		oldIDs[i] = id
	}
	newIDs := make([]int, m)
	for i, ln := range newLines {
		key := string(ln.base[ln.start:ln.end])
		id, ok := idOf[key]
		if !ok {
			id = nextID
			idOf[key] = id
			nextID++
		}
		newIDs[i] = id
	}

	maxDist := n + m
	offset := maxDist
	trace := make([][]int, 0, maxDist+1)

	Vprev := make([]int, 2*maxDist+1)
	for i := range Vprev {
		Vprev[i] = -1
	}

	x0 := 0
	y0 := 0
	for x0 < n && y0 < m && oldIDs[x0] == newIDs[y0] {
		x0++
		y0++
	}
	Vprev[offset+0] = x0
	trace = append(trace, append([]int(nil), Vprev...))

	found := x0 >= n && y0 >= m

	for D := 1; D <= maxDist && !found; D++ {
		V := make([]int, 2*maxDist+1)
		for i := range V {
			V[i] = -1
		}

		for k := -D; k <= D; k += 2 {
			var x int
			if k == -D || (k != D && Vprev[offset+(k-1)] < Vprev[offset+(k+1)]) {
				x = Vprev[offset+(k+1)]
			} else {
				x = Vprev[offset+(k-1)] + 1
			}
			y := x - k

			for x < n && y < m && oldIDs[x] == newIDs[y] {
				x++
				y++
			}
			V[offset+k] = x

			if x >= n && y >= m {
				trace = append(trace, V)
				found = true
				break
			}
		}

		if !found {
			trace = append(trace, V)
			Vprev = V
		}
	}

	type edit struct {
		kind    ChunkKind
		lineref lineRef
	}
	revEdits := make([]edit, 0, n+m)

	x := n
	y := m
	for D := len(trace) - 1; D >= 0; D-- {
		k := x - y

		var (
			prevK int
			prevX int
			prevY int
		)
		if D > 0 {
			prevV := trace[D-1]
			if k == -D || (k != D && prevV[offset+(k-1)] < prevV[offset+(k+1)]) {
				prevK = k + 1
			} else {
				prevK = k - 1
			}
			prevX = prevV[offset+prevK]
			prevY = prevX - prevK
		}

		for x > prevX && y > prevY {
			x--
			y--
			revEdits = append(revEdits, edit{kind: ChunkKindUnchanged, lineref: oldLines[x]})
		}

		if D == 0 {
			break
		}

		if x == prevX {
			y--
			revEdits = append(revEdits, edit{kind: ChunkKindAdded, lineref: newLines[y]})
		} else {
			x--
			revEdits = append(revEdits, edit{kind: ChunkKindDeleted, lineref: oldLines[x]})
		}
	}

	for i, j := 0, len(revEdits)-1; i < j; i, j = i+1, j-1 {
		revEdits[i], revEdits[j] = revEdits[j], revEdits[i]
	}

	var out []Chunk
	type meta struct {
		base  []byte
		start int
		end   int
	}
	var metas []meta

	for _, e := range revEdits {
		curBase := e.lineref.base
		curStart := e.lineref.start
		curEnd := e.lineref.end

		if len(out) == 0 || out[len(out)-1].Kind != e.kind {
			out = append(out, Chunk{Kind: e.kind, Data: curBase[curStart:curEnd]})
			metas = append(metas, meta{base: curBase, start: curStart, end: curEnd})
			continue
		}

		lastIdx := len(out) - 1
		lastMeta := metas[lastIdx]

		if bytes.Equal(lastMeta.base, curBase) && lastMeta.end == curStart {
			metas[lastIdx].end = curEnd
			out[lastIdx].Data = curBase[metas[lastIdx].start:metas[lastIdx].end]
			continue
		}

		out[lastIdx].Data = append(out[lastIdx].Data, curBase[curStart:curEnd]...)
		metas[lastIdx] = meta{base: nil, start: 0, end: 0}
	}

	return out, nil
}

// Chunk represents a contiguous region of lines categorized
// as unchanged, deleted, or added.
type Chunk struct {
	Kind ChunkKind
	Data []byte
}

// ChunkKind enumerates the type of diff chunk.
type ChunkKind int

const (
	// ChunkKindUnchanged represents an unchanged diff chunk.
	ChunkKindUnchanged ChunkKind = iota
	// ChunkKindDeleted represents a deleted diff chunk.
	ChunkKindDeleted
	// ChunkKindAdded represents an added diff chunk.
	ChunkKindAdded
)
