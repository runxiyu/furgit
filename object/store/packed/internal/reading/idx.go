package reading

import (
	"os"

	objectid "lindenii.org/go/furgit/object/id"
)

// idxFile stores one mapped and validated idx v2 file.
type idxFile struct {
	// idxName is the basename of this .idx file.
	idxName string
	// packName is the matching .pack basename.
	packName string
	// algo is the hash algorithm encoded by the index.
	algo objectid.Algorithm

	// file is the opened index file descriptor.
	file *os.File
	// data is the mapped index bytes.
	data []byte

	// fanout stores fanout table values.
	fanout [256]uint32
	// numObjects equals fanout[255].
	numObjects int

	// namesOffset starts the sorted object-id table.
	namesOffset int
	// offset32Offset starts the 32-bit offset table.
	offset32Offset int
	// offset64Offset starts the 64-bit offset table.
	offset64Offset int
	// offset64Count is the number of 64-bit offset entries.
	offset64Count int
}
