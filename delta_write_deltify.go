package furgit

import (
	"bytes"

	"github.com/cespare/xxhash/v2"
)

const (
	deltaMinChunk   = 32
	deltaMaxChunk   = 8192
	deltaSplitMask  = (1 << 8) - 1
	deltaStretchMax = (1 << 24) - 1
)

var deltaGearTab = [...]uint32{
	0x67ed26b7, 0x32da500c, 0x53d0fee0, 0xce387dc7, 0xcd406d90, 0x2e83a4d4,
	0x9fc9a38d, 0xb67259dc, 0xca6b1722, 0x6d2ea08c, 0x235cea2e, 0x3149bb5f,
	0x1beda787, 0x2a6b77d5, 0x2f22d9ac, 0x91fc0544, 0xe413acfa, 0x5a30ff7a,
	0xad6fdde0, 0x444fd0f5, 0x7ad87864, 0x58c5ff05, 0x8d2ec336, 0x2371f853,
	0x550f8572, 0x6aa448dd, 0x7c9ddbcf, 0x95221e14, 0x2a82ec33, 0xcbec5a78,
	0xc6795a0d, 0x243995b7, 0x1c909a2f, 0x4fded51c, 0x635d334b, 0x0e2b9999,
	0x2702968d, 0x856de1d5, 0x3325d60e, 0xeb6a7502, 0xec2a9844, 0x0905835a,
	0xa1820375, 0xa4be5cab, 0x96a6c058, 0x2c2ccd70, 0xba40fce3, 0xd794c46b,
	0x8fbae83e, 0xc3aa7899, 0x3d3ff8ed, 0xa0d42b5b, 0x571c0c97, 0xd2811516,
	0xf7e7b96c, 0x4fd2fcbd, 0xe2fdec94, 0x282cc436, 0x78e8e95c, 0x80a3b613,
	0xcfbee20c, 0xd4a32d1c, 0x2a12ff13, 0x6af82936, 0xe5630258, 0x8efa6a98,
	0x294fb2d1, 0xdeb57086, 0x5f0fddb3, 0xeceda7ce, 0x4c87305f, 0x3a6d3307,
	0xe22d2942, 0x9d060217, 0x1e42ed02, 0xb6f63b52, 0x4367f39f, 0x055cf262,
	0x03a461b2, 0x5ef9e382, 0x386bc03a, 0x2a1e79c7, 0xf1a0058b, 0xd4d2dea9,
	0x56baf37d, 0x5daff6cc, 0xf03a951d, 0xaef7de45, 0xa8f4581e, 0x3960b555,
	0xffbfff6d, 0xbe702a23, 0x8f5b6d6f, 0x061739fb, 0x98696f47, 0x3fd596d4,
	0x151eac6b, 0xa9fcc4f5, 0x69181a12, 0x3ac5a107, 0xb5198fe7, 0x96bcb1da,
	0x1b5ddf8e, 0xc757d650, 0x65865c3a, 0x8fc0a41a, 0x87435536, 0x99eda6f2,
	0x41874794, 0x29cff4e8, 0xb70efd9a, 0x3103f6e7, 0x84d2453b, 0x15a450bd,
	0x74f49af1, 0x60f664b1, 0xa1c86935, 0xfdafbce1, 0xe36353e3, 0x5d9ba739,
	0xbc0559ba, 0x708b0054, 0xd41d808c, 0xb2f31723, 0x9027c41f, 0xf136d165,
	0xb5374b12, 0x9420a6ac, 0x273958b6, 0xe6c2fad0, 0xebdc1f21, 0xfb33af8b,
	0xc71c25cd, 0xe9a2d8e5, 0xbeb38a50, 0xbceb7cc2, 0x4e4e73f0, 0xcd6c251d,
	0xde4c032c, 0x4b04ac30, 0x725b8b21, 0x4eb8c33b, 0x20d07b75, 0x0567aa63,
	0xb56b2bb7, 0xc1f5fd3a, 0xcafd35ca, 0x470dd4da, 0xfe4f94cd, 0xfb8de424,
	0xe8dbcf40, 0xfe50a37a, 0x62db5b5d, 0xf32f4ab6, 0x2c4a8a51, 0x18473dc0,
	0xfe0cbb6e, 0xfe399efd, 0xdf34ecc9, 0x6ccd5055, 0x46097073, 0x139135c2,
	0x721c76f6, 0x1c6a94b4, 0x6eee014d, 0x8a508e02, 0x3da538f5, 0x280d394f,
	0x5248a0c4, 0x3ce94c6c, 0x9a71ad3a, 0x8493dd05, 0xe43f0ab6, 0x18e4ed42,
	0x6c5c0e09, 0x42b06ec9, 0x8d330343, 0xa45b6f59, 0x2a573c0c, 0xd7fd3de6,
	0xeedeab68, 0x5c84dafc, 0xbbd1b1a8, 0xa3ce1ad1, 0x85b70bed, 0xb6add07f,
	0xa531309c, 0x8f8ab852, 0x564de332, 0xeac9ed0c, 0x73da402c, 0x3ec52761,
	0x43af2f4d, 0xd6ff45c8, 0x4c367462, 0xd553bd6a, 0x44724855, 0x3b2aa728,
	0x56e5eb65, 0xeaf16173, 0x33fa42ff, 0xd714bb5d, 0xfbd0a3b9, 0xaf517134,
	0x9416c8cd, 0x534cf94f, 0x548947c2, 0x34193569, 0x32f4389a, 0xfe7028bc,
	0xed73b1ed, 0x9db95770, 0x468e3922, 0x0440c3cd, 0x60059a62, 0x33504562,
	0x2b229fbd, 0x5174dca5, 0xf7028752, 0xd63c6aa8, 0x31276f38, 0x0646721c,
	0xb0191da8, 0xe00e6de0, 0x9eac1a6e, 0x9f7628a5, 0xed6c06ea, 0x0bb8af15,
	0xf119fb12, 0x38693c1c, 0x732bc0fe, 0x84953275, 0xb82ec888, 0x33a4f1b3,
	0x3099835e, 0x028a8782, 0x5fdd51d7, 0xc6c717b3, 0xb06caf71, 0x17c8c111,
	0x61bad754, 0x9fd03061, 0xe09df1af, 0x3bc9eb73, 0x85878413, 0x9889aaf2,
	0x3f5a9e46, 0x42c9f01f, 0x9984a4f4, 0xd5de43cc, 0xd294daed, 0xbecba2d2,
	0xf1f6e72c, 0x5551128a, 0x83af87e2, 0x6f0342ba,
}

type deltaBlock struct {
	length int
	offset int
	hash   uint64
}

type deltaTable struct {
	blocks  []deltaBlock
	nblocks int
	nalloc  int

	offs []uint32
	len  int
	size int
}

type deltaInstruction struct {
	copy   bool
	offset int
	length int
	data   []byte
}

func deltify(base, target []byte, seed uint64) ([]deltaInstruction, error) {
	if base == nil {
		base = []byte{}
	}
	if target == nil {
		target = []byte{}
	}
	dt, err := deltaTableInitMem(base, 0, len(base), seed)
	if err != nil {
		return nil, err
	}
	return deltifyMemMem(target, 0, len(target), seed, dt, base, 0, len(base))
}

func deltifyMemMem(target []byte, fileoffset, filesize int, seed uint64, dt *deltaTable, base []byte, baseOffset0, baseSize int) ([]deltaInstruction, error) {
	const allocChunkSize = 64
	instr := make([]deltaInstruction, 0, allocChunkSize)
	offset0 := fileoffset

	for fileoffset < filesize {
		blocklen := deltaNextBlockLen(target, fileoffset, filesize)
		if blocklen == 0 {
			if fileoffset < filesize {
				data := target[fileoffset:filesize]
				instr = append(instr, deltaInstruction{
					copy:   false,
					offset: fileoffset - offset0,
					length: len(data),
					data:   data,
				})
			}
			break
		}

		block := dt.lookupBlock(target[fileoffset:fileoffset+blocklen], blocklen, seed, base, baseOffset0)
		if block != nil {
			if err := deltaStretchMemMem(base, baseOffset0, baseSize, block, target, fileoffset+blocklen, filesize, &blocklen); err != nil {
				return nil, err
			}
			instr = append(instr, deltaInstruction{
				copy:   true,
				offset: block.offset,
				length: blocklen,
			})
		} else {
			data := target[fileoffset : fileoffset+blocklen]
			instr = append(instr, deltaInstruction{
				copy:   false,
				offset: fileoffset - offset0,
				length: len(data),
				data:   data,
			})
		}
		fileoffset += blocklen
	}

	return instr, nil
}

func deltaTableInitMem(data []byte, fileoffset, filesize int, seed uint64) (*deltaTable, error) {
	dt := &deltaTable{
		nalloc: 128,
		size:   128,
	}
	dt.blocks = make([]deltaBlock, dt.nalloc)
	dt.offs = make([]uint32, dt.size)

	offset0 := fileoffset
	for fileoffset < filesize {
		blocklen := deltaNextBlockLen(data, fileoffset, filesize)
		if blocklen == 0 {
			break
		}
		h := deltaHashBlock(data[fileoffset:fileoffset+blocklen], seed)
		if err := dt.addBlock(data, offset0, blocklen, fileoffset-offset0, h); err != nil {
			return nil, err
		}
		fileoffset += blocklen
	}
	return dt, nil
}

func deltaHashBlock(p []byte, seed uint64) uint64 {
	if seed == 0 {
		return xxhash.Sum64(p)
	}
	d := xxhash.NewWithSeed(seed)
	_, _ = d.Write(p)
	return d.Sum64()
}

func deltaNextBlockLen(data []byte, fileoffset, filesize int) int {
	if fileoffset >= filesize || filesize-fileoffset < deltaMinChunk {
		return 0
	}
	p := fileoffset + deltaMinChunk
	end := fileoffset + deltaMaxChunk
	if end > filesize {
		end = filesize
	}
	var gh uint32
	for p < end {
		gh = (gh << 1) + deltaGearTab[data[p]]
		p++
		if (gh & deltaSplitMask) == 0 {
			break
		}
	}
	return p - fileoffset
}

func (dt *deltaTable) addBlock(data []byte, offset0, length, offset int, h uint64) error {
	if length == 0 {
		return nil
	}
	if err := dt.resize(); err != nil {
		return err
	}
	idx, found := dt.lookupSlot(data, offset0, offset, length, h)
	if found {
		return nil
	}
	dt.offs[idx] = uint32(dt.nblocks + 1)
	dt.blocks[dt.nblocks] = deltaBlock{
		length: length,
		offset: offset,
		hash:   h,
	}
	dt.nblocks++
	dt.len++
	return nil
}

func (dt *deltaTable) lookupBlock(chunk []byte, length int, seed uint64, base []byte, baseOffset0 int) *deltaBlock {
	if dt == nil || dt.size == 0 {
		return nil
	}
	h := deltaHashBlock(chunk, seed)
	for i := int(h % uint64(dt.size)); dt.offs[i] != 0; i = (i + 1) % dt.size {
		block := &dt.blocks[dt.offs[i]-1]
		if block.hash != h || block.length != length {
			continue
		}
		baseOff := baseOffset0 + block.offset
		if baseOff < 0 || baseOff+length > len(base) {
			continue
		}
		if bytes.Equal(base[baseOff:baseOff+length], chunk) {
			return block
		}
	}
	return nil
}

func (dt *deltaTable) lookupSlot(data []byte, offset0, offset, length int, h uint64) (int, bool) {
	if dt == nil || dt.size == 0 {
		return 0, false
	}
	for i := int(h % uint64(dt.size)); ; i = (i + 1) % dt.size {
		if dt.offs[i] == 0 {
			return i, false
		}
		block := &dt.blocks[dt.offs[i]-1]
		if block.hash != h || block.length != length {
			continue
		}
		left := offset0 + block.offset
		right := offset0 + offset
		if left < 0 || right < 0 || left+length > len(data) || right+length > len(data) {
			continue
		}
		if bytes.Equal(data[left:left+length], data[right:right+length]) {
			return i, true
		}
	}
}

func (dt *deltaTable) resize() error {
	if dt.nblocks == dt.nalloc {
		newAlloc := dt.nalloc + 256
		newBlocks := make([]deltaBlock, newAlloc)
		copy(newBlocks, dt.blocks[:dt.nblocks])
		dt.blocks = newBlocks
		dt.nalloc = newAlloc
	}

	if dt.offs == nil || (dt.len*100)/dt.size > 70 {
		oldSize := dt.size
		dt.size *= 2
		if dt.size == 0 {
			dt.size = oldSize
			return ErrInvalidObject
		}
		dt.offs = make([]uint32, dt.size)
		for i := 0; i < dt.nblocks; i++ {
			block := &dt.blocks[i]
			idx := int(block.hash % uint64(dt.size))
			for dt.offs[idx] != 0 {
				idx = (idx + 1) % dt.size
			}
			dt.offs[idx] = uint32(i + 1)
		}
	}

	return nil
}

func deltaStretchMemMem(base []byte, baseOffset0, baseSize int, block *deltaBlock, data []byte, fileoffset, filesize int, blocklen *int) error {
	baseOffset := baseOffset0 + block.offset + *blocklen
	if baseOffset > baseSize {
		return ErrInvalidObject
	}
	if fileoffset > filesize {
		return ErrInvalidObject
	}
	maxlen := baseSize - baseOffset
	if avail := filesize - fileoffset; avail < maxlen {
		maxlen = avail
	}
	for i := 0; i < maxlen && *blocklen < deltaStretchMax-1; i++ {
		if data[fileoffset+i] != base[baseOffset+i] {
			break
		}
		*blocklen++
	}
	return nil
}
