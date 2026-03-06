package commitquery

// Marks returns the mark bits of one internal node.
func (ctx *Context) Marks(idx NodeIndex) markBits {
	return ctx.nodes[idx].marks
}

// HasAnyMarks reports whether one internal node has any requested bit.
func (ctx *Context) HasAnyMarks(idx NodeIndex, bits markBits) bool {
	return ctx.nodes[idx].marks&bits != 0
}

// HasAllMarks reports whether one internal node already has all requested bits.
func (ctx *Context) HasAllMarks(idx NodeIndex, bits markBits) bool {
	return ctx.nodes[idx].marks&bits == bits
}

// SetMarks ORs one set of mark bits into one internal node.
func (ctx *Context) SetMarks(idx NodeIndex, bits markBits) {
	newBits := bits &^ ctx.nodes[idx].marks
	if newBits == 0 {
		return
	}

	ctx.trackTouched(idx)
	ctx.nodes[idx].marks |= bits
}

// ClearMarks removes one set of mark bits from one internal node.
func (ctx *Context) ClearMarks(idx NodeIndex, bits markBits) {
	if ctx.nodes[idx].marks&bits == 0 {
		return
	}

	ctx.trackTouched(idx)
	ctx.nodes[idx].marks &^= bits
}

// BeginMarkPhase starts one tracked mark-mutation phase.
func (ctx *Context) BeginMarkPhase() {
	ctx.markPhase++
	if ctx.markPhase == 0 {
		ctx.markPhase++
		for i := range ctx.nodes {
			ctx.nodes[i].touchedPhase = 0
		}
	}

	ctx.touched = ctx.touched[:0]
}

// ClearTouchedMarks clears the provided bits from all nodes touched in the
// current mark phase.
func (ctx *Context) ClearTouchedMarks(bits markBits) {
	for _, idx := range ctx.touched {
		ctx.nodes[idx].marks &^= bits
	}
}

func (ctx *Context) trackTouched(idx NodeIndex) {
	if ctx.nodes[idx].touchedPhase == ctx.markPhase {
		return
	}

	ctx.nodes[idx].touchedPhase = ctx.markPhase
	ctx.touched = append(ctx.touched, idx)
}
