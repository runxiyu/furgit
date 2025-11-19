package furgit

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"syscall"

	"git.sr.ht/~runxiyu/furgit/internal/zlib"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
)

const (
	packMagic    = 0x5041434b
	packVersion2 = 2
)

type packlocation struct {
	PackPath string
	Offset   uint64
}

func (repo *Repository) packRead(id Hash) (StoredObject, error) {
	loc, err := repo.packIndexFind(id)
	if err != nil {
		return nil, err
	}
	return repo.packReadAt(loc, id)
}

func (repo *Repository) packIndexFind(id Hash) (packlocation, error) {
	midx, err := repo.multiPackIndex()
	if err == nil {
		loc, err := midx.lookup(id)
		if err == nil {
			return loc, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return packlocation{}, err
		}
	} else if !errors.Is(err, ErrNotFound) {
		return packlocation{}, err
	}

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

func (repo *Repository) packReadAt(loc packlocation, want Hash) (StoredObject, error) {
	ty, body, err := repo.packBodyResolveAtLocation(loc)
	if err != nil {
		return nil, err
	}
	data := body.Bytes()
	// if !repo.verifyTypedObject(ty, data, want) {
	// 	body.Release()
	// 	return nil, ErrInvalidObject
	// }
	obj, err := parseObjectBody(ty, want, data, repo)
	body.Release()
	return obj, err
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

func (repo *Repository) packTypeSizeByID(id Hash, seen map[packKey]struct{}) (ObjectType, int64, error) {
	loc, err := repo.packIndexFind(id)
	if err == nil {
		return repo.packTypeSizeAtLocation(loc, seen)
	}
	if !errors.Is(err, ErrNotFound) {
		return ObjectTypeInvalid, 0, err
	}
	return repo.looseTypeSize(id)
}

func packHeaderRead(r io.Reader) (ObjectType, int, error) {
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	ty := ObjectType((b[0] >> 4) & 0x07)
	size := int(b[0] & 0x0f)
	shift := 4
	for (b[0] & 0x80) != 0 {
		_, err = io.ReadFull(r, b[:])
		if err != nil {
			return ObjectTypeInvalid, 0, err
		}
		size |= int(b[0]&0x7f) << shift
		shift += 7
		if (b[0] & 0x80) == 0 {
			break
		}
	}
	return ty, size, nil
}

func packSectionInflate(pf *packFile, objectOfs uint64, r io.Reader, sizeHint int) (bufpool.Buffer, error) {
	if pf != nil {
		if br, ok := r.(*bytes.Reader); ok {
			total := br.Size()
			remaining := int64(br.Len())
			consumed := total - remaining
			start := objectOfs + uint64(consumed)
			if int64(consumed) < 0 || start > uint64(len(pf.data)) {
				return bufpool.Buffer{}, ErrInvalidObject
			}
			body, err := zlib.Decompress(pf.data[start:])
			if err != nil {
				return bufpool.Buffer{}, err
			}
			if sizeHint > 0 && len(body.Bytes()) != sizeHint {
				body.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			return body, nil
		}
	}

	zr, err := zlib.NewReader(r)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	defer func() { _ = zr.Close() }()

	body := bufpool.Borrow(bufpool.DefaultBufferCap)
	var scratch [32 * 1024]byte
	for {
		n, err := zr.Read(scratch[:])
		if n > 0 {
			body.Append(scratch[:n])
		}
		if err == io.EOF {
			if sizeHint > 0 && len(body.Bytes()) != sizeHint {
				body.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			return body, nil
		}
		if err != nil {
			body.Release()
			return bufpool.Buffer{}, err
		}
	}
}

func packDeltaReadOfsDistance(r io.Reader) (uint64, error) {
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return 0, err
	}
	dist := uint64(b[0] & 0x7f)
	for (b[0] & 0x80) != 0 {
		_, err = io.ReadFull(r, b[:])
		if err != nil {
			return 0, err
		}
		dist = ((dist + 1) << 7) + uint64(b[0]&0x7f)
	}
	return dist, nil
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

		r, err := pf.cursor(ofs)
		if err != nil {
			return ObjectTypeInvalid, 0, err
		}
		ty, size, err := packHeaderRead(r)
		if err != nil {
			return ObjectTypeInvalid, 0, err
		}
		if firstHeader {
			declaredSize = int64(size)
			firstHeader = false
		}

		switch ty {
		case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
			return ty, declaredSize, nil
		case ObjectTypeRefDelta:
			var base Hash
			_, err := io.ReadFull(r, base.data[:repo.hashSize])
			if err != nil {
				return ObjectTypeInvalid, 0, err
			}
			base.size = repo.hashSize
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
			dist, err := packDeltaReadOfsDistance(r)
			if err != nil {
				return ObjectTypeInvalid, 0, err
			}
			if ofs <= dist {
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
		r, err := pf.cursor(ofs)
		if err != nil {
			return fail(err)
		}
		ty, size, err := packHeaderRead(r)
		if err != nil {
			return fail(err)
		}

		switch ty {
		case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
			body, err = packSectionInflate(pf, ofs, r, size)
			if err != nil {
				return fail(err)
			}
			bodyReady = true
			resultTy = ty
			resolved = true
		case ObjectTypeRefDelta:
			var base Hash
			_, err := io.ReadFull(r, base.data[:repo.hashSize])
			if err != nil {
				return fail(err)
			}
			base.size = repo.hashSize
			delta, err := packSectionInflate(pf, ofs, r, 0)
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
			dist, err := packDeltaReadOfsDistance(r)
			if err != nil {
				return fail(err)
			}
			if ofs <= dist {
				return fail(ErrInvalidObject)
			}
			delta, err := packSectionInflate(pf, ofs, r, 0)
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

	header := make([]byte, 12)
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

	err = syscall.Madvise(region, syscall.MADV_RANDOM)
	if err != nil {
		_ = syscall.Munmap(region)
		return nil, err
	}
	err = syscall.Madvise(region, syscall.MADV_WILLNEED)
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

func (pf *packFile) cursor(ofs uint64) (io.Reader, error) {
	if pf == nil {
		return nil, ErrInvalidObject
	}
	if pf.size < 0 {
		return nil, ErrInvalidObject
	}
	if ofs > uint64(pf.size) {
		return nil, fmt.Errorf("furgit: pack: offset %d beyond %s", ofs, pf.relPath)
	}
	if ofs > uint64(math.MaxInt64) {
		return nil, fmt.Errorf("furgit: pack: offset %d too large", ofs)
	}
	return bytes.NewReader(pf.data[ofs:]), nil
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
