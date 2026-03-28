package packed

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"

	packfmt "codeberg.org/lindenii/furgit/format/packfile"
	"codeberg.org/lindenii/furgit/internal/intconv"
)

// packFile stores one mapped and validated .pack file.
type packFile struct {
	// name is the .pack basename.
	name string
	// file is the opened pack file descriptor.
	file *os.File
	// data is the mapped pack bytes.
	data []byte
}

// openPackFile maps and validates one pack file.
func openPackFile(name string, file *os.File, size int64) (*packFile, error) {
	if size < 12 {
		return nil, fmt.Errorf("objectstore/packed: pack %q too short", name)
	}

	if size > int64(int(^uint(0)>>1)) {
		return nil, fmt.Errorf("objectstore/packed: pack %q has unsupported size", name)
	}

	fd, err := intconv.UintptrToInt(file.Fd())
	if err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(fd, 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}

	if binary.BigEndian.Uint32(data[:4]) != packfmt.Signature {
		_ = syscall.Munmap(data)

		return nil, fmt.Errorf("objectstore/packed: pack %q invalid signature", name)
	}

	version := binary.BigEndian.Uint32(data[4:8])
	if !packfmt.VersionSupported(version) {
		_ = syscall.Munmap(data)

		return nil, fmt.Errorf("objectstore/packed: pack %q unsupported version %d", name, version)
	}

	return &packFile{name: name, file: file, data: data}, nil
}

// close unmaps and closes one pack handle.
func (pack *packFile) close() error {
	var closeErr error

	if pack.data != nil {
		err := syscall.Munmap(pack.data)
		if err != nil && closeErr == nil {
			closeErr = err
		}

		pack.data = nil
	}

	if pack.file != nil {
		err := pack.file.Close()
		if err != nil && closeErr == nil {
			closeErr = err
		}

		pack.file = nil
	}

	return closeErr
}
