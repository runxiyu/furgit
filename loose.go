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

func loosePath(id Hash) string {
	hex := id.String()
	return filepath.Join("objects", hex[:2], hex[2:])
}

func (repo *Repository) looseRead(id Hash) (Object, error) {
	ty, body, err := repo.looseReadTyped(id)
	if err != nil {
		return nil, err
	}
	return parseObjectBody(ty, id, body)
}

func (repo *Repository) looseReadTyped(id Hash) (ObjType, []byte, error) {
	path := repo.repoPath(loosePath(id))
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

	space := bytes.IndexByte(header, ' ')
	if space < 0 {
		return ObjInvalid, nil, ErrInvalidObject
	}
	tyStr := string(header[:space])
	var ty ObjType
	switch tyStr {
	case "blob":
		ty = ObjBlob
	case "tree":
		ty = ObjTree
	case "commit":
		ty = ObjCommit
	case "tag":
		ty = ObjTag
	default:
		return ObjInvalid, nil, ErrInvalidObject
	}
	expect := header[space+1:]
	size, err := strconv.Atoi(string(expect))
	if err != nil {
		return ObjInvalid, nil, fmt.Errorf("furgit: loose: size parse: %w", err)
	}
	if size != len(body) {
		return ObjInvalid, nil, ErrInvalidObject
	}
	if !verifyRawObject(raw, id) {
		return ObjInvalid, nil, ErrInvalidObject
	}

	out := append([]byte(nil), body...)
	return ty, out, nil
}
