package furgit

import (
	"bytes"
	stdzlib "compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
	"git.sr.ht/~runxiyu/furgit/internal/zlib"
)

const looseHeaderLimit = 4096

// loosePath returns the path for a loose object, validating hash size.
func (repo *Repository) loosePath(id Hash) (string, error) {
	if id.size != repo.hashSize {
		return "", fmt.Errorf("furgit: hash size mismatch: got %d, expected %d", id.size, repo.hashSize)
	}
	hex := id.String()
	return filepath.Join("objects", hex[:2], hex[2:]), nil
}

func (repo *Repository) looseRead(id Hash) (StoredObject, error) {
	ty, body, err := repo.looseReadTyped(id)
	if err != nil {
		return nil, err
	}
	obj, err := parseObjectBody(ty, id, body.Bytes(), repo)
	body.Release()
	return obj, err
}

func (repo *Repository) looseReadTyped(id Hash) (ObjectType, bufpool.Buffer, error) {
	path, err := repo.loosePath(id)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	path = repo.repoPath(path)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ObjectTypeInvalid, bufpool.Buffer{}, ErrNotFound
		}
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	defer func() { _ = f.Close() }()

	compressed, err := io.ReadAll(f)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}

	raw, err := zlib.Decompress(compressed)
	if err != nil {
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}

	rawBytes := raw.Bytes()
	nul := bytes.IndexByte(rawBytes, 0)
	if nul < 0 {
		raw.Release()
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	}

	header := rawBytes[:nul]
	body := rawBytes[nul+1:]

	ty, declaredSize, err := parseLooseHeader(header)
	if err != nil {
		raw.Release()
		return ObjectTypeInvalid, bufpool.Buffer{}, err
	}
	if declaredSize != int64(len(body)) {
		raw.Release()
		return ObjectTypeInvalid, bufpool.Buffer{}, ErrInvalidObject
	}

	copy(rawBytes, body)
	raw.Resize(len(body))
	// if !repo.verifyRawObject(raw, id) {
	// 	return ObjectTypeInvalid, nil, ErrInvalidObject
	// }

	return ty, raw, nil
}

func (repo *Repository) looseTypeSize(id Hash) (ObjectType, int64, error) {
	path, err := repo.loosePath(id)
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	path = repo.repoPath(path)
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ObjectTypeInvalid, 0, ErrNotFound
		}
		return ObjectTypeInvalid, 0, err
	}
	defer func() { _ = f.Close() }()

	zr, err := stdzlib.NewReader(f)
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	defer func() { _ = zr.Close() }()

	header := make([]byte, 0, 64)
	chunk := make([]byte, 128)
	for {
		n, readErr := zr.Read(chunk)
		if n > 0 {
			data := chunk[:n]
			if nul := bytes.IndexByte(data, 0); nul >= 0 {
				header = append(header, data[:nul]...)
				if len(header) > looseHeaderLimit {
					return ObjectTypeInvalid, 0, ErrInvalidObject
				}
				break
			}
			header = append(header, data...)
			if len(header) > looseHeaderLimit {
				return ObjectTypeInvalid, 0, ErrInvalidObject
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return ObjectTypeInvalid, 0, ErrInvalidObject
			}
			return ObjectTypeInvalid, 0, readErr
		}
	}
	return parseLooseHeader(header)
}

func parseLooseHeader(header []byte) (ObjectType, int64, error) {
	space := bytes.IndexByte(header, ' ')
	if space < 0 {
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
	ty, err := objTypeFromName(string(header[:space]))
	if err != nil {
		return ObjectTypeInvalid, 0, err
	}
	expect := header[space+1:]
	if len(expect) == 0 {
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
	size, err := strconv.ParseInt(string(expect), 10, 64)
	if err != nil {
		return ObjectTypeInvalid, 0, fmt.Errorf("furgit: loose: size parse: %w", err)
	}
	if size < 0 {
		return ObjectTypeInvalid, 0, ErrInvalidObject
	}
	return ty, size, nil
}

func objTypeFromName(name string) (ObjectType, error) {
	switch name {
	case objectTypeNameBlob:
		return ObjectTypeBlob, nil
	case objectTypeNameTree:
		return ObjectTypeTree, nil
	case objectTypeNameCommit:
		return ObjectTypeCommit, nil
	case objectTypeNameTag:
		return ObjectTypeTag, nil
	default:
		return ObjectTypeInvalid, ErrInvalidObject
	}
}

// WriteLooseObject writes an object to the repository as a loose object.
func (repo *Repository) WriteLooseObject(obj Object) (Hash, error) {
	var raw []byte
	var err error

	switch o := obj.(type) {
	case *Blob:
		raw, err = o.Serialize()
	case *Tree:
		raw, err = o.Serialize()
	case *Commit:
		raw, err = o.Serialize()
	case *Tag:
		raw, err = o.Serialize()
	default:
		return Hash{}, fmt.Errorf("furgit: unsupported object type for writing: %T", obj)
	}
	// TODO: Consider adding serialize to the interface?

	if err != nil {
		return Hash{}, err
	}

	id := repo.computeRawHash(raw)
	path, err := repo.loosePath(id)
	if err != nil {
		return Hash{}, err
	}
	path = repo.repoPath(path)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Hash{}, err
	}

	var buf bytes.Buffer
	zw := stdzlib.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		return Hash{}, err
	}
	if err := zw.Close(); err != nil {
		return Hash{}, err
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return Hash{}, err
	}

	return id, nil
}
