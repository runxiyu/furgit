package furgit

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const looseHeaderLimit = 4096

// loosePath returns the path for a loose object, validating hash size.
func (repo *Repository) loosePath(id Hash) (string, error) {
	if id.size != repo.HashSize {
		return "", fmt.Errorf("furgit: hash size mismatch: got %d, expected %d", id.size, repo.HashSize)
	}
	hex := id.String()
	return filepath.Join("objects", hex[:2], hex[2:]), nil
}

func (repo *Repository) looseRead(id Hash) (Object, error) {
	ty, body, err := repo.looseReadTyped(id)
	if err != nil {
		return nil, err
	}
	return parseObjectBody(ty, id, body, repo)
}

func (repo *Repository) looseReadTyped(id Hash) (ObjType, []byte, error) {
	path, err := repo.loosePath(id)
	if err != nil {
		return ObjInvalid, nil, err
	}
	path = repo.repoPath(path)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ObjInvalid, nil, ErrNotFound
		}
		return ObjInvalid, nil, err
	}
	defer func() { _ = f.Close() }()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return ObjInvalid, nil, err
	}
	defer func() { _ = zr.Close() }()

	raw, err := io.ReadAll(zr)
	if err != nil {
		return ObjInvalid, nil, err
	}

	nul := bytes.IndexByte(raw, 0)
	if nul < 0 {
		return ObjInvalid, nil, ErrInvalidObject
	}

	header := raw[:nul]
	body := raw[nul+1:]

	ty, declaredSize, err := parseLooseHeader(header)
	if err != nil {
		return ObjInvalid, nil, err
	}
	if declaredSize != int64(len(body)) {
		return ObjInvalid, nil, ErrInvalidObject
	}
	if !repo.verifyRawObject(raw, id) {
		return ObjInvalid, nil, ErrInvalidObject
	}

	out := append([]byte(nil), body...)
	return ty, out, nil
}

func (repo *Repository) looseTypeSize(id Hash) (ObjType, int64, error) {
	path, err := repo.loosePath(id)
	if err != nil {
		return ObjInvalid, 0, err
	}
	path = repo.repoPath(path)
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ObjInvalid, 0, ErrNotFound
		}
		return ObjInvalid, 0, err
	}
	defer func() { _ = f.Close() }()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return ObjInvalid, 0, err
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
					return ObjInvalid, 0, ErrInvalidObject
				}
				break
			}
			header = append(header, data...)
			if len(header) > looseHeaderLimit {
				return ObjInvalid, 0, ErrInvalidObject
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return ObjInvalid, 0, ErrInvalidObject
			}
			return ObjInvalid, 0, readErr
		}
	}
	return parseLooseHeader(header)
}

func parseLooseHeader(header []byte) (ObjType, int64, error) {
	space := bytes.IndexByte(header, ' ')
	if space < 0 {
		return ObjInvalid, 0, ErrInvalidObject
	}
	ty, err := objTypeFromName(string(header[:space]))
	if err != nil {
		return ObjInvalid, 0, err
	}
	expect := header[space+1:]
	if len(expect) == 0 {
		return ObjInvalid, 0, ErrInvalidObject
	}
	size, err := strconv.ParseInt(string(expect), 10, 64)
	if err != nil {
		return ObjInvalid, 0, fmt.Errorf("furgit: loose: size parse: %w", err)
	}
	if size < 0 {
		return ObjInvalid, 0, ErrInvalidObject
	}
	return ty, size, nil
}

func objTypeFromName(name string) (ObjType, error) {
	switch name {
	case objNameBlob:
		return ObjBlob, nil
	case objNameTree:
		return ObjTree, nil
	case objNameCommit:
		return ObjCommit, nil
	case objNameTag:
		return ObjTag, nil
	default:
		return ObjInvalid, ErrInvalidObject
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
	zw := zlib.NewWriter(&buf)
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
