package read

import (
	"os"
	"syscall"

	"codeberg.org/lindenii/furgit/format/commitgraph"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

func openLayer(root *os.Root, relPath string, algo objectid.Algorithm) (*layer, error) {
	file, err := root.Open(relPath)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	size := info.Size()
	if size < int64(commitgraph.HeaderSize+commitgraph.FanoutSize+algo.Size()) {
		_ = file.Close()

		return nil, &MalformedError{Path: relPath, Reason: "file too short"}
	}

	mapLen, err := intconv.Int64ToUint64(size)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	mapLenInt, err := intconv.Uint64ToInt(mapLen)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	fd, err := intconv.UintptrToInt(file.Fd())
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	data, err := syscall.Mmap(fd, 0, mapLenInt, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		_ = file.Close()

		return nil, err
	}

	out := &layer{
		path: relPath,
		file: file,
		data: data,
	}

	parseErr := parseLayer(out, algo)
	if parseErr != nil {
		_ = out.close()

		return nil, parseErr
	}

	verifyErr := verifyTrailerHash(out.data, algo, relPath)
	if verifyErr != nil {
		_ = out.close()

		return nil, verifyErr
	}

	return out, nil
}
