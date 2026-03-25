package packed

import "syscall"

// close unmaps and closes one idx handle.
func (index *idxFile) close() error {
	var closeErr error

	if index.data != nil {
		err := syscall.Munmap(index.data)
		if err != nil && closeErr == nil {
			closeErr = err
		}

		index.data = nil
	}

	if index.file != nil {
		err := index.file.Close()
		if err != nil && closeErr == nil {
			closeErr = err
		}

		index.file = nil
	}

	return closeErr
}
