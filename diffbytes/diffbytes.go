package diffbytes

import "bytes"

// DiffBytes performs a line-based diff.
// Lines are bytes up to and including '\n' (final line may lack '\n').
func DiffBytes(oldB, newB []byte) ([]BytesDiffChunk, error) {
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
		key := bytesToString(ln.base[ln.start:ln.end])
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
		key := bytesToString(ln.base[ln.start:ln.end])
		id, ok := idOf[key]
		if !ok {
			id = nextID
			idOf[key] = id
			nextID++
		}
		newIDs[i] = id
	}

	max := n + m
	offset := max
	trace := make([][]int, 0, max+1)

	Vprev := make([]int, 2*max+1)
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

	for D := 1; D <= max && !found; D++ {
		V := make([]int, 2*max+1)
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
		kind    BytesDiffChunkKind
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
			revEdits = append(revEdits, edit{kind: BytesDiffChunkKindUnchanged, lineref: oldLines[x]})
		}

		if D == 0 {
			break
		}

		if x == prevX {
			y--
			revEdits = append(revEdits, edit{kind: BytesDiffChunkKindAdded, lineref: newLines[y]})
		} else {
			x--
			revEdits = append(revEdits, edit{kind: BytesDiffChunkKindDeleted, lineref: oldLines[x]})
		}
	}

	for i, j := 0, len(revEdits)-1; i < j; i, j = i+1, j-1 {
		revEdits[i], revEdits[j] = revEdits[j], revEdits[i]
	}

	var out []BytesDiffChunk
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
			out = append(out, BytesDiffChunk{Kind: e.kind, Data: curBase[curStart:curEnd]})
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

// BytesDiffChunk represents a contiguous region of bytes categorized
// as unchanged, deleted, or added.
type BytesDiffChunk struct {
	Kind BytesDiffChunkKind
	Data []byte
}

// BytesDiffChunkKind enumerates the type of diff chunk.
type BytesDiffChunkKind int

const (
	// BytesDiffChunkKindUnchanged represents an unchanged diff chunk.
	BytesDiffChunkKindUnchanged BytesDiffChunkKind = iota
	// BytesDiffChunkKindDeleted represents a deleted diff chunk.
	BytesDiffChunkKindDeleted
	// BytesDiffChunkKindAdded represents an added diff chunk.
	BytesDiffChunkKindAdded
)
