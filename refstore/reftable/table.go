package reftable

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"syscall"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
)

const (
	reftableMagic = "REFT"

	version1 = 1
	version2 = 2

	blockTypeRef   = byte('r')
	blockTypeIndex = byte('i')
)

var (
	hashIDSHA1   = binary.BigEndian.Uint32([]byte("sha1"))
	hashIDSHA256 = binary.BigEndian.Uint32([]byte("s256"))
)

// tableFile is one opened and mapped reftable file.
type tableFile struct {
	// name is the table filename from tables.list.
	name string
	// algo is the expected object ID algorithm.
	algo objectid.Algorithm

	// file is the opened table file.
	file *os.File
	// data is the mapped table bytes.
	data []byte

	// headerLen is 24 for v1 or 28 for v2.
	headerLen int
	// blockSize is configured alignment; 0 means unaligned.
	blockSize int

	// refEnd is the exclusive end of ref blocks section.
	refEnd int
	// refIndexPos is the root ref-index block position, or 0 when absent.
	refIndexPos uint64
}

// recordValue is one decoded reference record value.
type recordValue struct {
	// deleted marks a tombstone record.
	deleted bool
	// detachedID stores a direct object ID for detached refs.
	detachedID objectid.ObjectID
	// hasDetached reports whether detachedID is valid.
	hasDetached bool
	// peeled stores an optional peeled ID for annotated tags.
	peeled *objectid.ObjectID
	// symbolicTarget stores symref target for symbolic refs.
	symbolicTarget string
}

// openTableFile maps and validates one reftable file.
func openTableFile(root *os.Root, name string, algo objectid.Algorithm) (*tableFile, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	size := info.Size()
	if size < 0 || size > int64(int(^uint(0)>>1)) {
		_ = file.Close()
		return nil, fmt.Errorf("refstore/reftable: table %q has unsupported size", name)
	}
	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	out := &tableFile{name: name, algo: algo, file: file, data: data}
	if err := out.parseMeta(); err != nil {
		_ = out.close()
		return nil, err
	}
	return out, nil
}

// close unmaps and closes one table file.
func (table *tableFile) close() error {
	var closeErr error
	if table.data != nil {
		if err := syscall.Munmap(table.data); err != nil && closeErr == nil {
			closeErr = err
		}
		table.data = nil
	}
	if table.file != nil {
		if err := table.file.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		table.file = nil
	}
	return closeErr
}

// parseMeta validates header/footer and section boundaries.
func (table *tableFile) parseMeta() error {
	if len(table.data) < 24 {
		return fmt.Errorf("refstore/reftable: table %q: file too short", table.name)
	}
	if string(table.data[:4]) != reftableMagic {
		return fmt.Errorf("refstore/reftable: table %q: bad magic", table.name)
	}
	version := table.data[4]
	switch version {
	case version1:
		table.headerLen = 24
		if table.algo != objectid.AlgorithmSHA1 {
			return fmt.Errorf("refstore/reftable: table %q: version 1 requires sha1", table.name)
		}
	case version2:
		table.headerLen = 28
		if len(table.data) < table.headerLen {
			return fmt.Errorf("refstore/reftable: table %q: truncated header", table.name)
		}
		hashID := binary.BigEndian.Uint32(table.data[24:28])
		if err := validateHashID(hashID, table.algo); err != nil {
			return fmt.Errorf("refstore/reftable: table %q: %w", table.name, err)
		}
	default:
		return fmt.Errorf("refstore/reftable: table %q: unsupported version %d", table.name, version)
	}
	table.blockSize = int(readUint24(table.data[5:8]))

	footerLen := 68
	if version == version2 {
		footerLen = 72
	}
	if len(table.data) < footerLen {
		return fmt.Errorf("refstore/reftable: table %q: missing footer", table.name)
	}
	footerStart := len(table.data) - footerLen
	footer := table.data[footerStart:]
	if string(footer[:4]) != reftableMagic || footer[4] != version {
		return fmt.Errorf("refstore/reftable: table %q: invalid footer header", table.name)
	}
	wantCRC := binary.BigEndian.Uint32(footer[footerLen-4:])
	haveCRC := crc32.ChecksumIEEE(footer[:footerLen-4])
	if wantCRC != haveCRC {
		return fmt.Errorf("refstore/reftable: table %q: footer crc mismatch", table.name)
	}
	if version == version2 {
		hashID := binary.BigEndian.Uint32(footer[24:28])
		if err := validateHashID(hashID, table.algo); err != nil {
			return fmt.Errorf("refstore/reftable: table %q: %w", table.name, err)
		}
	}

	off := table.headerLen
	table.refIndexPos = binary.BigEndian.Uint64(footer[off : off+8])
	off += 8
	objPosAndLen := binary.BigEndian.Uint64(footer[off : off+8])
	off += 8
	objPos := objPosAndLen >> 5
	objIndexPos := binary.BigEndian.Uint64(footer[off : off+8])
	off += 8
	logPos := binary.BigEndian.Uint64(footer[off : off+8])
	off += 8
	logIndexPos := binary.BigEndian.Uint64(footer[off : off+8])
	_ = objIndexPos
	_ = logIndexPos

	refEnd := uint64(footerStart)
	if table.refIndexPos != 0 && table.refIndexPos < refEnd {
		refEnd = table.refIndexPos
	}
	if objPos != 0 && objPos < refEnd {
		refEnd = objPos
	}
	if logPos != 0 && logPos < refEnd {
		refEnd = logPos
	}
	if refEnd < uint64(table.headerLen) || refEnd > uint64(len(table.data)) {
		return fmt.Errorf("refstore/reftable: table %q: invalid ref section", table.name)
	}
	if table.refIndexPos > uint64(len(table.data)) {
		return fmt.Errorf("refstore/reftable: table %q: invalid ref index position", table.name)
	}
	table.refEnd = int(refEnd)
	return nil
}

// validateHashID validates a reftable v2 hash identifier.
func validateHashID(hashID uint32, algo objectid.Algorithm) error {
	switch hashID {
	case hashIDSHA1:
		if algo != objectid.AlgorithmSHA1 {
			return errors.New("hash id sha1 mismatch")
		}
		return nil
	case hashIDSHA256:
		if algo != objectid.AlgorithmSHA256 {
			return errors.New("hash id s256 mismatch")
		}
		return nil
	default:
		return fmt.Errorf("unknown hash id 0x%08x", hashID)
	}
}

// toRef converts a decoded record value into a public ref value.
func (record recordValue) toRef(name string) (ref.Ref, error) {
	if record.deleted {
		return nil, errors.New("refstore/reftable: cannot materialize deleted record")
	}
	if record.symbolicTarget != "" {
		return ref.Symbolic{RefName: name, Target: record.symbolicTarget}, nil
	}
	if !record.hasDetached {
		return nil, errors.New("refstore/reftable: malformed detached record")
	}
	return ref.Detached{RefName: name, ID: record.detachedID, Peeled: record.peeled}, nil
}
