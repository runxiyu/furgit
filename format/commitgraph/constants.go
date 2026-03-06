package commitgraph

const (
	fileSignature = 0x43475048 // "CGPH"
	fileVersion   = 1
)

const (
	chunkOIDF = 0x4f494446 // "OIDF"
	chunkOIDL = 0x4f49444c // "OIDL"
	chunkCDAT = 0x43444154 // "CDAT"
	chunkGDA2 = 0x47444132 // "GDA2"
	chunkGDO2 = 0x47444f32 // "GDO2"
	chunkEDGE = 0x45444745 // "EDGE"
	chunkBIDX = 0x42494458 // "BIDX"
	chunkBDAT = 0x42444154 // "BDAT"
	chunkBASE = 0x42415345 // "BASE"
)

const (
	headerSize     = 8
	chunkEntrySize = 12
	fanoutSize     = 256 * 4
)

const (
	parentNone      = 0x70000000
	parentExtraMask = 0x80000000
	parentLastMask  = 0x7fffffff

	generationOverflow = 0x80000000
)
