package packed

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

const packSignature = 0x5041434b

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
	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint32(data[:4]) != packSignature {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("objectstore/packed: pack %q invalid signature", name)
	}
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 && version != 3 {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("objectstore/packed: pack %q unsupported version %d", name, version)
	}
	return &packFile{name: name, file: file, data: data}, nil
}

// close unmaps and closes one pack handle.
func (pack *packFile) close() error {
	var closeErr error
	if pack.data != nil {
		if err := syscall.Munmap(pack.data); err != nil && closeErr == nil {
			closeErr = err
		}
		pack.data = nil
	}
	if pack.file != nil {
		if err := pack.file.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		pack.file = nil
	}
	return closeErr
}
