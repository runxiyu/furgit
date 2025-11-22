package flatex

// windowDecoder implements the sliding window used in decompression.
type windowDecoder struct {
	hist []byte

	wrPos int
	rdPos int
	full  bool
}

func (wd *windowDecoder) init(size int) {
	*wd = windowDecoder{hist: wd.hist}

	if cap(wd.hist) < size {
		wd.hist = make([]byte, size)
	}
	wd.hist = wd.hist[:size]

	wd.wrPos = 0
	wd.rdPos = 0
	wd.full = false
}

func (wd *windowDecoder) histSize() int {
	if wd.full {
		return len(wd.hist)
	}
	return wd.wrPos
}

func (wd *windowDecoder) availRead() int {
	return wd.wrPos - wd.rdPos
}

func (wd *windowDecoder) availWrite() int {
	return len(wd.hist) - wd.wrPos
}

func (wd *windowDecoder) writeSlice() []byte {
	return wd.hist[wd.wrPos:]
}

func (wd *windowDecoder) writeMark(cnt int) {
	wd.wrPos += cnt
}

func (wd *windowDecoder) writeByte(c byte) {
	wd.hist[wd.wrPos] = c
	wd.wrPos++
}

func (wd *windowDecoder) writeCopy(dist, length int) int {
	dstBase := wd.wrPos
	dstPos := dstBase
	srcPos := dstPos - dist
	endPos := dstPos + length
	if endPos > len(wd.hist) {
		endPos = len(wd.hist)
	}

	if srcPos < 0 {
		srcPos += len(wd.hist)
		dstPos += copy(wd.hist[dstPos:endPos], wd.hist[srcPos:])
		srcPos = 0
	}

	for dstPos < endPos {
		dstPos += copy(wd.hist[dstPos:endPos], wd.hist[srcPos:dstPos])
	}

	wd.wrPos = dstPos
	return dstPos - dstBase
}

func (wd *windowDecoder) tryWriteCopy(dist, length int) int {
	dstPos := wd.wrPos
	endPos := dstPos + length
	if dstPos < dist || endPos > len(wd.hist) {
		return 0
	}
	dstBase := dstPos
	srcPos := dstPos - dist

	for dstPos < endPos {
		dstPos += copy(wd.hist[dstPos:endPos], wd.hist[srcPos:dstPos])
	}

	wd.wrPos = dstPos
	return dstPos - dstBase
}

func (wd *windowDecoder) readFlush() []byte {
	toRead := wd.hist[wd.rdPos:wd.wrPos]
	wd.rdPos = wd.wrPos
	if wd.wrPos == len(wd.hist) {
		wd.wrPos, wd.rdPos = 0, 0
		wd.full = true
	}
	return toRead
}
