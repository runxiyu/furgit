//go:build unix

package mmap

import (
	"fmt"
	"os"
	"syscall"

	"lindenii.org/go/lgo/intconv"
)

// Mmap is one read-only memory mapping of an entire file.
//
// The mapping remains valid
// after the originating file is closed.
//
// Truncating the mapped file can cause
// a fatal SIGBUS on later access;
// callers should only map files
// that are never truncated while in use.
//
// Labels: Close-Caller.
type Mmap struct {
	// data is the mapped region,
	// or nil for empty files and closed mappings.
	data []byte
}

// Open maps file read-only in its entirety.
//
// file is only used during Open
// and may be closed once Open returns.
func Open(file *os.File) (*Mmap, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("internal/mmap: %w", err)
	}

	size64, err := intconv.Int64ToUint64(info.Size())
	if err != nil {
		return nil, fmt.Errorf("internal/mmap: %w", err)
	}

	size, err := intconv.Uint64ToInt(size64)
	if err != nil {
		return nil, fmt.Errorf("internal/mmap: %w", err)
	}

	if size == 0 {
		return &Mmap{data: nil}, nil
	}

	fd, err := intconv.UintptrToInt(file.Fd())
	if err != nil {
		return nil, fmt.Errorf("internal/mmap: %w", err)
	}

	data, err := syscall.Mmap(fd, 0, size, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, fmt.Errorf("internal/mmap: %w", err)
	}

	return &Mmap{data: data}, nil
}

// Data returns the mapped bytes.
//
// Labels: Life-Parent, Mut-No.
func (mmap *Mmap) Data() []byte {
	return mmap.data
}

// Close unmaps the mapping,
// invalidating previously returned data.
//
// Labels: Idem-Yes, MT-Unsafe.
func (mmap *Mmap) Close() error {
	if mmap.data == nil {
		return nil
	}

	data := mmap.data
	mmap.data = nil

	err := syscall.Munmap(data)
	if err != nil {
		return fmt.Errorf("internal/mmap: %w", err)
	}

	return nil
}
