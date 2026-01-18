package furgit

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"
	"syscall"

	"codeberg.org/lindenii/furgit/internal/bufpool"
	"codeberg.org/lindenii/furgit/internal/zlibx"
)

const (
	packMagic    = 0x5041434b
	packVersion2 = 2
)

type packlocation struct {
	PackPath string
	Offset   uint64
}

func (repo *Repository) packRead(id Hash) (ObjectType, bufpool.Buffer, error) {
	loc, err := repo.packIndexFind(id)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	return repo.packReadAt(loc, id)
}

func (repo *Repository) packIndexFind(id Hash) (packlocation, error) {
	idxs, err := repo.packIndexes()
	if err != nil {
		return packlocation{}, err
	}
	for _, idx := range idxs {
		loc, err := idx.lookup(id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return packlocation{}, err
		}
		return loc, nil
	}
	return packlocation{}, ErrNotFound
}

func (repo *Repository) packReadAt(loc packlocation, want Hash) (ObjectType, bufpool.Buffer, error) {
	ty, body, err := repo.packBodyResolveAtLocation(loc)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	return ty, body, nil
}

func (repo *Repository) packBodyResolveAtLocation(loc packlocation) (ObjectType, bufpool.Buffer, error) {
	pf, err := repo.packFile(loc.PackPath)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	return repo.packBodyResolveWithin(pf, loc.Offset)
}

func (repo *Repository) packTypeSizeAtLocation(loc packlocation, seen map[packKey]struct{}) (ObjectType, int64, error) {
	pf, err := repo.packFile(loc.PackPath)
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	return repo.packTypeSizeWithin(pf, loc.Offset, seen)
}

func packHeaderParse(data []byte) (ObjectType, int, int, error) {
	if len(data) == 0 {
		return ObjectTypeInvalid, 0, 0, io.ErrUnexpectedEOF
	}
	b := data[0]
	ty := ObjectType((b >> 4) & 0x07)
	size := int(b & 0x0f)
	shift := 4
	consumed := 1
	for (b & 0x80) != 0 {
		if consumed >= len(data) {
			return ObjectTypeInvalid, 0, 0, io.ErrUnexpectedEOF
		}
		b = data[consumed]
		size |= int(b&0x7f) << shift
		shift += 7
		consumed++
	}
	return ty, size, consumed, nil
}

func packSectionInflate(pf *packFile, start uint64, sizeHint int) (bufpool.Buffer, error) {
	if start > uint64(len(pf.data)) {
		return bufpool.Buffer{}, ErrInvalidObject
	}
	body, err := zlibx.DecompressSized(pf.data[start:], sizeHint)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	if sizeHint > 0 && len(body.Bytes()) != sizeHint {
		body.Release()
		return bufpool.Buffer{}, ErrInvalidObject
	}
	return body, nil
}

func packDeltaReadOfsDistance(data []byte) (uint64, int, error) {
	if len(data) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b := data[0]
	dist := uint64(b & 0x7f)
	consumed := 1
	for (b & 0x80) != 0 {
		if consumed >= len(data) {
			return 0, 0, io.ErrUnexpectedEOF
		}
		b = data[consumed]
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}
	return dist, consumed, nil
}

type packKey struct {
	path string
	ofs  uint64
}

func (repo *Repository) packTypeSizeWithin(pf *packFile, ofs uint64, seen map[packKey]struct{}) (ObjectType, int64, error) {
	if pf == nil {
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
	if seen == nil {
		seen = make(map[packKey]struct{})
	}
	var visited []packKey
	defer func() {
		for _, key := range visited {
			delete(seen, key)
		}
	}()

	var declaredSize int64
	firstHeader := true

	for {
		key := packKey{path: pf.relPath, ofs: ofs}
		if _, dup := seen[key]; dup {
			return ObjectTypeInvalid, 0, ErrInvalidObject
		}
		seen[key] = struct{}{}
		visited = append(visited, key)

		if ofs >= uint64(len(pf.data)) {
			return ObjectTypeInvalid, 0, ErrInvalidObject
		}
		ty, size, consumed, err := packHeaderParse(pf.data[ofs:])
		if err != nil {
			return ObjectTypeInvalid, 0, err
		}
		if firstHeader {
			declaredSize = int64(size)
			firstHeader = false
		}

		if uint64(consumed) > uint64(len(pf.data))-ofs {
			return ObjectTypeInvalid, 0, io.ErrUnexpectedEOF
		}
		dataStart := ofs + uint64(consumed)
		switch ty {
		case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
			return ty, declaredSize, nil
		case ObjectTypeRefDelta:
			hashEnd := dataStart + uint64(repo.hashAlgo.size())
			if hashEnd > uint64(len(pf.data)) {
				return ObjectTypeInvalid, 0, io.ErrUnexpectedEOF
			}
			var base Hash
			copy(base.data[:], pf.data[dataStart:hashEnd])
			base.algo = repo.hashAlgo
			loc, err := repo.packIndexFind(base)
			if err == nil {
				pf, err = repo.packFile(loc.PackPath)
				if err != nil {
					return ObjectTypeInvalid, 0, err
				}
				ofs = loc.Offset
				continue
			}
			if !errors.Is(err, ErrNotFound) {
				return ObjectTypeInvalid, 0, err
			}
			baseTy, _, err := repo.looseTypeSize(base)
			if err != nil {
				return ObjectTypeInvalid, 0, err
			}
			return baseTy, declaredSize, nil
		case ObjectTypeOfsDelta:
			dist, distConsumed, err := packDeltaReadOfsDistance(pf.data[dataStart:])
			if err != nil {
				return ObjectTypeInvalid, 0, err
			}
			if ofs <= dist {
				return ObjectTypeInvalid, 0, ErrInvalidObject
			}
			dataStart += uint64(distConsumed)
			if dataStart > uint64(len(pf.data)) {
				return ObjectTypeInvalid, 0, ErrInvalidObject
			}
			ofs -= dist
		case ObjectTypeInvalid, ObjectTypeFuture:
			return ObjectTypeInvalid, 0, ErrInvalidObject
		default:
			return ObjectTypeInvalid, 0, ErrInvalidObject
		}
	}
}

func (repo *Repository) packBodyResolveWithin(pf *packFile, ofs uint64) (ObjectType, bufpool.Buffer, error) {
	if pf == nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	}

	type deltaFrame struct {
		delta bufpool.Buffer
	}
	var frames []deltaFrame
	defer func() {
		for i := range frames {
			frames[i].delta.Release()
		}
	}()

	var (
		body      bufpool.Buffer
		bodyReady bool
		resultTy  ObjectType
	)
	fail := func(err error) (ObjectType, bufpool.Buffer, error) {
		if bodyReady {
			body.Release()
			bodyReady = false
		}
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}

	resolved := false
	for !resolved {
		if ofs >= uint64(len(pf.data)) {
			return fail(ErrInvalidObject)
		}
		ty, size, consumed, err := packHeaderParse(pf.data[ofs:])
		if err != nil {
			return fail(err)
		}
		if uint64(consumed) > uint64(len(pf.data))-ofs {
			return fail(io.ErrUnexpectedEOF)
		}
		dataStart := ofs + uint64(consumed)

		switch ty {
		case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
			body, err = packSectionInflate(pf, dataStart, size)
			if err != nil {
				return fail(err)
			}
			bodyReady = true
			resultTy = ty
			resolved = true
		case ObjectTypeRefDelta:
			hashEnd := dataStart + uint64(repo.hashAlgo.size())
			if hashEnd > uint64(len(pf.data)) {
				return fail(io.ErrUnexpectedEOF)
			}
			var base Hash
			copy(base.data[:], pf.data[dataStart:hashEnd])
			base.algo = repo.hashAlgo
			delta, err := packSectionInflate(pf, hashEnd, 0)
			if err != nil {
				return fail(err)
			}
			frames = append(frames, deltaFrame{delta: delta})

			loc, err := repo.packIndexFind(base)
			if err == nil {
				pf, err = repo.packFile(loc.PackPath)
				if err != nil {
					return fail(err)
				}
				ofs = loc.Offset
				continue
			}
			if !errors.Is(err, ErrNotFound) {
				return fail(err)
			}
			resultTy, body, err = repo.looseReadTyped(base)
			if err != nil {
				return fail(err)
			}
			bodyReady = true
			resolved = true
		case ObjectTypeOfsDelta:
			dist, distConsumed, err := packDeltaReadOfsDistance(pf.data[dataStart:])
			if err != nil {
				return fail(err)
			}
			if ofs <= dist {
				return fail(ErrInvalidObject)
			}
			deltaStart := dataStart + uint64(distConsumed)
			if deltaStart > uint64(len(pf.data)) {
				return fail(ErrInvalidObject)
			}
			delta, err := packSectionInflate(pf, deltaStart, 0)
			if err != nil {
				return fail(err)
			}
			frames = append(frames, deltaFrame{delta: delta})
			ofs -= dist
		case ObjectTypeInvalid, ObjectTypeFuture:
			return fail(ErrInvalidObject)
		default:
			return fail(ErrInvalidObject)
		}
	}

	for i := len(frames) - 1; i >= 0; i-- {
		out, err := packDeltaApply(body, frames[i].delta)
		body.Release()
		bodyReady = false
		frames[i].delta.Release()
		if err != nil {
			return fail(err)
		}
		body = out
		bodyReady = true
	}
	frames = nil
	return resultTy, body, nil
}

func packDeltaApply(base, delta bufpool.Buffer) (bufpool.Buffer, error) {
	pos := 0
	baseBytes := base.Bytes()
	deltaBytes := delta.Bytes()
	srcSize, err := packVarintRead(deltaBytes, &pos)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	dstSize, err := packVarintRead(deltaBytes, &pos)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	if srcSize != len(baseBytes) {
		return bufpool.Buffer{}, ErrInvalidObject
	}
	out := bufpool.Borrow(dstSize)
	out.Resize(dstSize)
	outBytes := out.Bytes()
	outPos := 0

	for pos < len(deltaBytes) {
		op := deltaBytes[pos]
		pos++
		switch {
		case op&0x80 != 0:
			off := 0
			n := 0
			if op&0x01 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos])
				pos++
			}
			if op&0x02 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 8
				pos++
			}
			if op&0x04 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 16
				pos++
			}
			if op&0x08 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 24
				pos++
			}
			if op&0x10 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos])
				pos++
			}
			if op&0x20 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos]) << 8
				pos++
			}
			if op&0x40 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos]) << 16
				pos++
			}
			if n == 0 {
				n = 0x10000
			}
			if off+n > len(baseBytes) || outPos+n > len(outBytes) {
				out.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			copy(outBytes[outPos:], baseBytes[off:off+n])
			outPos += n
		case op != 0:
			n := int(op)
			if pos+n > len(deltaBytes) || outPos+n > len(outBytes) {
				out.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			copy(outBytes[outPos:], deltaBytes[pos:pos+n])
			pos += n
			outPos += n
		default:
			out.Release()
			return bufpool.Buffer{}, ErrInvalidObject
		}
	}

	if outPos != len(outBytes) {
		out.Release()
		return bufpool.Buffer{}, ErrInvalidObject
	}
	return out, nil
}

func packVarintRead(buf []byte, pos *int) (int, error) {
	res := 0
	shift := 0
	for {
		if *pos >= len(buf) {
			return 0, ErrInvalidObject
		}
		b := buf[*pos]
		*pos++
		res |= int(b&0x7f) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}
	return res, nil
}

type packFile struct {
	relPath string
	size    int64
	data    []byte
	closeMu sync.Once
}

func openPackFile(absPath, rel string) (*packFile, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.Size() < 12 {
		_ = f.Close()
		return nil, ErrInvalidObject
	}

	var headerArr [12]byte
	header := headerArr[:]
	_, err = io.ReadFull(f, header)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	magic := binary.BigEndian.Uint32(header[:4])
	ver := binary.BigEndian.Uint32(header[4:8])
	if magic != packMagic || ver != packVersion2 {
		_ = f.Close()
		return nil, ErrInvalidObject
	}

	region, err := syscall.Mmap(
		int(f.Fd()),
		0,
		int(stat.Size()),
		syscall.PROT_READ,
		syscall.MAP_PRIVATE,
	)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	err = f.Close()
	if err != nil {
		_ = syscall.Munmap(region)
		return nil, err
	}

	return &packFile{
		relPath: rel,
		size:    stat.Size(),
		data:    region,
	}, nil
}

func (pf *packFile) Close() error {
	if pf == nil {
		return nil
	}
	var closeErr error
	pf.closeMu.Do(func() {
		if len(pf.data) > 0 {
			if err := syscall.Munmap(pf.data); closeErr == nil {
				closeErr = err
			}
			pf.data = nil
		}
	})
	return closeErr
}

func (repo *Repository) packFile(rel string) (*packFile, error) {
	repo.packFilesMu.RLock()
	pf, ok := repo.packFiles[rel]
	repo.packFilesMu.RUnlock()
	if ok {
		return pf, nil
	}

	pf, err := openPackFile(repo.repoPath(rel), rel)
	if err != nil {
		return nil, err
	}

	repo.packFilesMu.Lock()
	if existing, ok := repo.packFiles[rel]; ok {
		repo.packFilesMu.Unlock()
		_ = pf.Close()
		return existing, nil
	}
	repo.packFiles[rel] = pf
	repo.packFilesMu.Unlock()
	return pf, nil
}
