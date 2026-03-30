package reading

import (
	"bufio"

	deltaapply "codeberg.org/lindenii/furgit/format/packfile/delta/apply"
)

// deltaDeclaredSizeAt returns the resolved object size declared by one delta
// stream header at dataOffset.
func deltaDeclaredSizeAt(pack *packFile, dataOffset int) (int64, error) {
	reader, err := zlibReaderAt(pack, dataOffset)
	if err != nil {
		return 0, err
	}

	defer func() { _ = reader.Close() }()

	br := bufio.NewReaderSize(reader, 32)

	_, size, err := deltaapply.ReadHeaderSizes(br)
	if err != nil {
		return 0, err
	}

	return int64(size), nil
}
