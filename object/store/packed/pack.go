package packed

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/internal/mmap"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

var (
	errPackTruncated       = errors.New("truncated")
	errPackMalformedHeader = errors.New("malformed header")
	errPackCountMismatch   = errors.New("object count differs from index")
	errPackTrailerMismatch = errors.New("trailer hash differs from index")
)

// pack is one discovered pack:
// its base name, its parsed index, and its mapped data.
// All fields are immutable after openPack.
type pack struct {
	// name is the pack base name, like "pack-<hash>".
	name string

	// idxMapping owns the mapped pack index bytes,
	// and idx is the parsed index view over them.
	idxMapping *mmap.Mmap
	idx        packidx.Packidx

	// dataMapping owns the mapped pack data bytes,
	// and data aliases them.
	dataMapping *mmap.Mmap
	data        []byte

	bloomMapping *mmap.Mmap
	filter       *bloom.Bloom
}

// openPack opens, maps, and validates
// one pack index and its pack data
// by pack base name.
func openPack(root *os.Root, name string, objectFormat id.ObjectFormat) (*pack, error) {
	idxMapping, err := mapFile(root, name+".idx")
	if err != nil {
		return nil, err
	}

	idx, err := packidx.Parse(idxMapping.Data(), objectFormat.Size())
	if err != nil {
		_ = idxMapping.Close()

		return nil, fmt.Errorf("%w: index %q: %w", ErrMalformedPackedStore, name, err)
	}

	dataMapping, err := mapFile(root, name+".pack")
	if err != nil {
		_ = idxMapping.Close()

		return nil, err
	}

	err = validatePackData(dataMapping.Data(), &idx, objectFormat.Size())
	if err != nil {
		_ = idxMapping.Close()
		_ = dataMapping.Close()

		return nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, name, err)
	}

	bloomMapping, filter := openBloom(root, name, objectFormat, idx.PackHash())

	return &pack{
		name:         name,
		idxMapping:   idxMapping,
		idx:          idx,
		dataMapping:  dataMapping,
		data:         dataMapping.Data(),
		bloomMapping: bloomMapping,
		filter:       filter,
	}, nil
}

func openBloom(root *os.Root, name string, objectFormat id.ObjectFormat, packHash []byte) (*mmap.Mmap, *bloom.Bloom) {
	mapping, err := mapFile(root, name+".bloom")
	if err != nil {
		return nil, nil
	}

	filter, err := bloom.Parse(mapping.Data(), objectFormat)
	if err != nil {
		_ = mapping.Close()

		return nil, nil
	}

	if !bytes.Equal(filter.PackHash(), packHash) {
		_ = mapping.Close()

		return nil, nil
	}

	return mapping, &filter
}

// mapFile opens and maps one file under root.
func mapFile(root *os.Root, name string) (*mmap.Mmap, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("object/store/packed: %w", err)
	}

	defer func() { _ = file.Close() }()

	mapping, err := mmap.Open(file)
	if err != nil {
		return nil, fmt.Errorf("object/store/packed: %q: %w", name, err)
	}

	return mapping, nil
}

// validatePackData checks one mapped pack
// against the pack format and its index.
func validatePackData(data []byte, idx *packidx.Packidx, hashSize int) error {
	if len(data) < packfile.HeaderLen+hashSize {
		return errPackTruncated
	}

	header, err := packfile.ParseHeader(data)
	if err != nil {
		return fmt.Errorf("%w: %w", errPackMalformedHeader, err)
	}

	count := uint64(header.ObjectCount)

	numObjects, err := intconv.IntToUint64(idx.NumObjects())
	if err != nil {
		return fmt.Errorf("object count: %w", err)
	}

	if count != numObjects {
		return errPackCountMismatch
	}

	if !bytes.Equal(data[len(data)-hashSize:], idx.PackHash()) {
		return errPackTrailerMismatch
	}

	return nil
}

// close releases the pack data, index, and filter mappings.
func (pack *pack) close() error {
	errs := []error{
		pack.dataMapping.Close(),
		pack.idxMapping.Close(),
	}

	if pack.bloomMapping != nil {
		errs = append(errs, pack.bloomMapping.Close())
	}

	return errors.Join(errs...)
}
