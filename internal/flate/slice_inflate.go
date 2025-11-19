package flate

import (
	"io"
	"math/bits"
)

// sliceInflater is a specialized DEFLATE decoder that reads directly from an
// in-memory byte slice. It mirrors the main decompressor but avoids the
// overhead of the Reader interfaces, enabling faster byte-slice decoding.
type sliceInflater struct {
	input   []byte
	pos     int
	roffset int64

	b  uint32
	nb uint

	h1, h2 huffmanDecoder

	bits     *[maxNumLit + maxNumDist]int
	codebits *[numCodes]int

	dict dictDecoder

	toRead    []byte
	step      func(*sliceInflater)
	stepState int
	final     bool
	err       error
	hl, hd    *huffmanDecoder
	copyLen   int
	copyDist  int
}

func (f *sliceInflater) reset(src []byte, dict []byte) error {
	bits := f.bits
	codebits := f.codebits
	dictState := f.dict
	*f = sliceInflater{
		input:    src,
		bits:     bits,
		codebits: codebits,
		dict:     dictState,
		step:     (*sliceInflater).nextBlock,
	}
	f.dict.init(maxMatchOffset, dict)
	return nil
}

func (f *sliceInflater) readByte() (byte, error) {
	if f.pos >= len(f.input) {
		return 0, io.ErrUnexpectedEOF
	}
	b := f.input[f.pos]
	f.pos++
	f.roffset++
	return b, nil
}

func (f *sliceInflater) readBytes(n int) ([]byte, error) {
	if n < 0 || f.pos+n > len(f.input) {
		f.pos = len(f.input)
		return nil, io.ErrUnexpectedEOF
	}
	s := f.input[f.pos : f.pos+n]
	f.pos += n
	f.roffset += int64(n)
	return s, nil
}

func (f *sliceInflater) nextBlock() {
	for f.nb < 1+2 {
		if err := f.moreBits(); err != nil {
			f.err = err
			return
		}
	}
	f.final = f.b&1 == 1
	f.b >>= 1
	typ := f.b & 3
	f.b >>= 2
	f.nb -= 1 + 2
	switch typ {
	case 0:
		f.dataBlock()
	case 1:
		f.hl = &fixedHuffmanDecoder
		f.hd = nil
		f.huffmanBlock()
	case 2:
		if err := f.readHuffman(); err != nil {
			f.err = err
			return
		}
		f.hl = &f.h1
		f.hd = &f.h2
		f.huffmanBlock()
	default:
		f.err = CorruptInputError(f.roffset)
	}
}

func (f *sliceInflater) huffmanBlock() {
	const (
		stateInit = iota
		stateDict
	)
	switch f.stepState {
	case stateInit:
		goto readLiteral
	case stateDict:
		goto copyHistory
	}

readLiteral:
	{
		v, err := f.huffSym(f.hl)
		if err != nil {
			f.err = err
			return
		}
		var n uint
		var length int
		switch {
		case v < 256:
			f.dict.writeByte(byte(v))
			if f.dict.availWrite() == 0 {
				f.toRead = f.dict.readFlush()
				f.step = (*sliceInflater).huffmanBlock
				f.stepState = stateInit
				return
			}
			goto readLiteral
		case v == 256:
			f.finishBlock()
			return
		case v < 265:
			length = v - (257 - 3)
			n = 0
		case v < 269:
			length = v*2 - (265*2 - 11)
			n = 1
		case v < 273:
			length = v*4 - (269*4 - 19)
			n = 2
		case v < 277:
			length = v*8 - (273*8 - 35)
			n = 3
		case v < 281:
			length = v*16 - (277*16 - 67)
			n = 4
		case v < 285:
			length = v*32 - (281*32 - 131)
			n = 5
		case v < maxNumLit:
			length = 258
			n = 0
		default:
			f.err = CorruptInputError(f.roffset)
			return
		}
		if n > 0 {
			for f.nb < n {
				if err = f.moreBits(); err != nil {
					f.err = err
					return
				}
			}
			length += int(f.b & uint32(1<<n-1))
			f.b >>= n
			f.nb -= n
		}

		var dist int
		if f.hd == nil {
			for f.nb < 5 {
				if err = f.moreBits(); err != nil {
					f.err = err
					return
				}
			}
			dist = int(bits.Reverse8(uint8(f.b & 0x1F << 3)))
			f.b >>= 5
			f.nb -= 5
		} else {
			if dist, err = f.huffSym(f.hd); err != nil {
				f.err = err
				return
			}
		}

		switch {
		case dist < 4:
			dist++
		case dist < maxNumDist:
			nb := uint(dist-2) >> 1
			extra := (dist & 1) << nb
			for f.nb < nb {
				if err = f.moreBits(); err != nil {
					f.err = err
					return
				}
			}
			extra |= int(f.b & uint32(1<<nb-1))
			f.b >>= nb
			f.nb -= nb
			dist = 1<<(nb+1) + 1 + extra
		default:
			f.err = CorruptInputError(f.roffset)
			return
		}

		if dist > f.dict.histSize() {
			f.err = CorruptInputError(f.roffset)
			return
		}

		f.copyLen, f.copyDist = length, dist
		goto copyHistory
	}

copyHistory:
	{
		cnt := f.dict.tryWriteCopy(f.copyDist, f.copyLen)
		if cnt == 0 {
			cnt = f.dict.writeCopy(f.copyDist, f.copyLen)
		}
		f.copyLen -= cnt

		if f.dict.availWrite() == 0 || f.copyLen > 0 {
			f.toRead = f.dict.readFlush()
			f.step = (*sliceInflater).huffmanBlock
			f.stepState = stateDict
			return
		}
		goto readLiteral
	}
}

func (f *sliceInflater) dataBlock() {
	f.nb = 0
	f.b = 0

	hdr, err := f.readBytes(4)
	if err != nil {
		f.err = err
		return
	}
	n := int(hdr[0]) | int(hdr[1])<<8
	nn := int(hdr[2]) | int(hdr[3])<<8
	if uint16(nn) != uint16(^n) {
		f.err = CorruptInputError(f.roffset)
		return
	}

	if n == 0 {
		f.toRead = f.dict.readFlush()
		f.finishBlock()
		return
	}

	f.copyLen = n
	f.copyData()
}

func (f *sliceInflater) copyData() {
	for {
		if f.copyLen == 0 {
			f.finishBlock()
			return
		}
		buf := f.dict.writeSlice()
		if len(buf) == 0 {
			f.toRead = f.dict.readFlush()
			f.step = (*sliceInflater).copyData
			return
		}
		n := f.copyLen
		if n > len(buf) {
			n = len(buf)
		}
		if f.pos+n > len(f.input) {
			f.err = io.ErrUnexpectedEOF
			return
		}
		copy(buf[:n], f.input[f.pos:f.pos+n])
		f.pos += n
		f.roffset += int64(n)
		f.copyLen -= n
		f.dict.writeMark(n)
		if f.dict.availWrite() == 0 {
			f.toRead = f.dict.readFlush()
			f.step = (*sliceInflater).copyData
			return
		}
	}
}

func (f *sliceInflater) finishBlock() {
	if f.final {
		if f.dict.availRead() > 0 {
			f.toRead = f.dict.readFlush()
		}
		f.err = io.EOF
	}
	f.step = (*sliceInflater).nextBlock
	f.stepState = 0
}

func (f *sliceInflater) moreBits() error {
	c, err := f.readByte()
	if err != nil {
		return err
	}
	f.b |= uint32(c) << (f.nb & 31)
	f.nb += 8
	return nil
}

func (f *sliceInflater) huffSym(h *huffmanDecoder) (int, error) {
	n := uint(h.min)
	nb, b := f.nb, f.b
	for {
		for nb < n {
			c, err := f.readByte()
			if err != nil {
				f.b = b
				f.nb = nb
				return 0, err
			}
			b |= uint32(c) << (nb & 31)
			nb += 8
		}
		chunk := h.chunks[b&(huffmanNumChunks-1)]
		n = uint(chunk & huffmanCountMask)
		if n > huffmanChunkBits {
			chunk = h.links[chunk>>huffmanValueShift][(b>>huffmanChunkBits)&h.linkMask]
			n = uint(chunk & huffmanCountMask)
		}
		if n <= nb {
			if n == 0 {
				f.b = b
				f.nb = nb
				f.err = CorruptInputError(f.roffset)
				return 0, f.err
			}
			f.b = b >> (n & 31)
			f.nb = nb - n
			return int(chunk >> huffmanValueShift), nil
		}
	}
}

func (f *sliceInflater) readHuffman() error {
	for f.nb < 5+5+4 {
		if err := f.moreBits(); err != nil {
			return err
		}
	}
	nlit := int(f.b&0x1F) + 257
	if nlit > maxNumLit {
		return CorruptInputError(f.roffset)
	}
	f.b >>= 5
	ndist := int(f.b&0x1F) + 1
	if ndist > maxNumDist {
		return CorruptInputError(f.roffset)
	}
	f.b >>= 5
	nclen := int(f.b&0xF) + 4
	f.b >>= 4
	f.nb -= 5 + 5 + 4
	codebits := f.codebits[:]
	bits := f.bits[:]
	for i := range codebits {
		codebits[i] = 0
	}
	for i := range bits {
		bits[i] = 0
	}
	for i := 0; i < nclen; i++ {
		for f.nb < 3 {
			if err := f.moreBits(); err != nil {
				return err
			}
		}
		codebits[codeOrder[i]] = int(f.b & 0x7)
		f.b >>= 3
		f.nb -= 3
	}
	if !f.h1.init(codebits) {
		return CorruptInputError(f.roffset)
	}
	for i := range bits {
		bits[i] = 0
	}
	i := 0
	for i < nlit+ndist {
		x, err := f.huffSym(&f.h1)
		if err != nil {
			return err
		}
		switch {
		case x < 16:
			bits[i] = x
			i++
		case x == 16:
			if i == 0 {
				return CorruptInputError(f.roffset)
			}
			repeat := 3
			for f.nb < 2 {
				if err := f.moreBits(); err != nil {
					return err
				}
			}
			repeat += int(f.b & 0x3)
			f.b >>= 2
			f.nb -= 2
			for repeat > 0 {
				if i >= len(bits) {
					return CorruptInputError(f.roffset)
				}
				bits[i] = bits[i-1]
				i++
				repeat--
			}
		case x == 17:
			repeat := 3
			for f.nb < 3 {
				if err := f.moreBits(); err != nil {
					return err
				}
			}
			repeat += int(f.b & 0x7)
			f.b >>= 3
			f.nb -= 3
			for repeat > 0 {
				if i >= len(bits) {
					return CorruptInputError(f.roffset)
				}
				bits[i] = 0
				i++
				repeat--
			}
		case x == 18:
			repeat := 11
			for f.nb < 7 {
				if err := f.moreBits(); err != nil {
					return err
				}
			}
			repeat += int(f.b & 0x7F)
			f.b >>= 7
			f.nb -= 7
			for repeat > 0 {
				if i >= len(bits) {
					return CorruptInputError(f.roffset)
				}
				bits[i] = 0
				i++
				repeat--
			}
		default:
			return CorruptInputError(f.roffset)
		}
	}
	if !f.h1.init(bits[:nlit]) {
		return CorruptInputError(f.roffset)
	}
	if !f.h2.init(bits[nlit : nlit+ndist]) {
		return CorruptInputError(f.roffset)
	}
	if f.h1.min < bits[endBlockMarker] {
		f.h1.min = bits[endBlockMarker]
	}
	return nil
}
