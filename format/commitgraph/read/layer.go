package read

import (
	"os"

	"codeberg.org/lindenii/furgit/format/commitgraph/bloom"
)

type layer struct {
	path       string
	file       *os.File
	data       []byte
	numCommits uint32
	baseCount  uint32
	globalFrom uint32

	chunkOIDFanout    []byte
	chunkOIDLookup    []byte
	chunkCommit       []byte
	chunkGeneration   []byte
	chunkGenerationOv []byte
	chunkExtraEdges   []byte
	chunkBloomIndex   []byte
	chunkBloomData    []byte
	chunkBaseGraphs   []byte

	bloomSettings *bloom.Settings
}
