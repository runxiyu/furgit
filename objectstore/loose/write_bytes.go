package loose

import (
	"compress/zlib"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

const tempObjectFilePrefix = "tmp_obj_"

// WriteBytesFull writes a full serialized object as "type size\\x00content".
func (store *Store) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	var zero objectid.ObjectID

	if _, _, err := parseRaw(raw); err != nil {
		return zero, err
	}

	id := store.algo.Sum(raw)
	relPath, err := store.objectPath(id)
	if err != nil {
		return zero, err
	}
	if err := store.writeCompressedAtomic(relPath, raw); err != nil {
		return zero, err
	}
	return id, nil
}

// WriteBytesContent writes typed content bytes as a loose object.
func (store *Store) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	var zero objectid.ObjectID

	header, ok := objectheader.Encode(ty, int64(len(content)))
	if !ok {
		return zero, fmt.Errorf("objectstore/loose: failed to encode object header for type %d", ty)
	}

	raw := make([]byte, len(header)+len(content))
	copy(raw, header)
	copy(raw[len(header):], content)
	return store.WriteBytesFull(raw)
}

// writeCompressedAtomic compresses raw and writes it to relPath atomically.
func (store *Store) writeCompressedAtomic(relPath string, raw []byte) error {
	if _, err := store.root.Stat(relPath); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	dir := filepath.Dir(relPath)
	if err := store.root.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpRelPath, tmpFile, err := store.createTempObjectFile(dir)
	if err != nil {
		return err
	}

	cleanup := true
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
		}
		if cleanup {
			_ = store.root.Remove(tmpRelPath)
		}
	}()

	zw := zlib.NewWriter(tmpFile)
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	tmpFile = nil

	if err := store.root.Rename(tmpRelPath, relPath); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}

	cleanup = false
	return nil
}

// createTempObjectFile creates a unique temporary object file within dir.
func (store *Store) createTempObjectFile(dir string) (string, *os.File, error) {
	for range 16 {
		relPath := filepath.Join(dir, tempObjectFilePrefix+rand.Text())
		file, err := store.root.OpenFile(relPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return relPath, file, nil
		}
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return "", nil, err
	}

	return "", nil, errors.New("objectstore/loose: failed to create temporary object file")
}
