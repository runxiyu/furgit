package furgit

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"syscall"

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
	if !repo.verifyTypedObject(ty, data, want) {
		body.Release()
		return nil, ErrInvalidObject
	}
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

func packSectionInflate(r io.Reader, sizeHint int) (bufpool.Buffer, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	defer func() { _ = zr.Close() }()

	if sizeHint > 0 {
		body := bufpool.Borrow(sizeHint)
		body.Resize(sizeHint)
		_, err := io.ReadFull(zr, body.Bytes())
		if err != nil {
			body.Release()
			return bufpool.Buffer{}, err
		}
		var extra [1]byte
		_, err = zr.Read(extra[:])
		if err != io.EOF {
			body.Release()
			if err == nil {
				return bufpool.Buffer{}, ErrInvalidObject
			}
			return bufpool.Buffer{}, err
		}
		return body, nil
	}

	body := bufpool.Borrow(bufpool.DefaultBufferCap)
	var scratch [32 * 1024]byte
	for {
		n, err := zr.Read(scratch[:])
		if n > 0 {
			body.Append(scratch[:n])
		}
		if err == io.EOF {
			return body, nil
		}
		if err != nil {
			body.Release()
			return bufpool.Buffer{}, err
		}
	}
}

func (repo *Repository) packDeltaResolveOfs(pf *packFile, deltaOffset uint64, r io.Reader) (ObjectType, bufpool.Buffer, error) {
	dist, err := packDeltaReadOfsDistance(r)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	var baseOfs uint64
	if deltaOffset > dist {
		baseOfs = deltaOffset - dist
	}
	if baseOfs == 0 {
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	}
	ty, body, err := repo.packBodyResolveWithin(pf, baseOfs)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	delta, err := packSectionInflate(r, 0)
	if err != nil {
		body.Release()
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	out, err := packDeltaApply(body, delta)
	delta.Release()
	body.Release()
	if err != nil {
		out.Release()
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	return ty, out, nil
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

func (repo *Repository) packBodyResolveByID(id Hash) (ObjectType, bufpool.Buffer, error) {
	loc, err := repo.packIndexFind(id)
	if err == nil {
		return repo.packBodyResolveAtLocation(loc)
	}
	if !errors.Is(err, ErrNotFound) {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	ty, body, err := repo.looseReadTyped(id)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	return ty, bufpool.FromOwned(body), nil
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
	key := packKey{path: pf.relPath, ofs: ofs}
	if _, dup := seen[key]; dup {
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
	seen[key] = struct{}{}
	defer delete(seen, key)

	r, err := pf.cursor(ofs)
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	ty, size, err := packHeaderRead(r)
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	declaredSize := int64(size)

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
		baseTy, _, err := repo.packTypeSizeByID(base, seen)
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
		baseOfs := ofs - dist
		baseTy, _, err := repo.packTypeSizeWithin(pf, baseOfs, seen)
		if err != nil {
			return ObjectTypeInvalid, 0, err
		}
		return baseTy, declaredSize, nil
	case ObjectTypeInvalid, ObjectTypeFuture:
		return ObjectTypeInvalid, 0, ErrInvalidObject
	default:
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
}

func (repo *Repository) packBodyResolveWithin(pf *packFile, ofs uint64) (ObjectType, bufpool.Buffer, error) {
	r, err := pf.cursor(ofs)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	ty, size, err := packHeaderRead(r)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}

	switch ty {
	case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
		body, err := packSectionInflate(r, size)
		return ty, body, err
	case ObjectTypeRefDelta:
		var base Hash
		_, err := io.ReadFull(r, base.data[:repo.hashSize])
		if err != nil {
			return ObjectTypeInvalid, bufpool.Buffer{}, err
		}
		base.size = repo.hashSize
		delta, err := packSectionInflate(r, 0)
		if err != nil {
			return ObjectTypeInvalid, bufpool.Buffer{}, err
		}
		bt, body, err := repo.packBodyResolveByID(base)
		if err != nil {
			delta.Release()
			return ObjectTypeInvalid, bufpool.Buffer{}, err
		}
		out, err := packDeltaApply(body, delta)
		delta.Release()
		body.Release()
		if err != nil {
			out.Release()
			return ObjectTypeInvalid, bufpool.Buffer{}, err
		}
		return bt, out, nil
	case ObjectTypeOfsDelta:
		return repo.packDeltaResolveOfs(pf, ofs, r)
	case ObjectTypeInvalid, ObjectTypeFuture:
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	default:
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	}
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
