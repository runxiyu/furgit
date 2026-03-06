package commitgraph

const (
	FileSignature = 0x43475048 // "CGPH"
	FileVersion   = 1
)

const (
	ChunkOIDF = 0x4f494446 // "OIDF"
	ChunkOIDL = 0x4f49444c // "OIDL"
	ChunkCDAT = 0x43444154 // "CDAT"
	ChunkGDA2 = 0x47444132 // "GDA2"
	ChunkGDO2 = 0x47444f32 // "GDO2"
	ChunkEDGE = 0x45444745 // "EDGE"
	ChunkBIDX = 0x42494458 // "BIDX"
	ChunkBDAT = 0x42444154 // "BDAT"
	ChunkBASE = 0x42415345 // "BASE"
)

const (
	HeaderSize     = 8
	ChunkEntrySize = 12
	FanoutSize     = 256 * 4
)

const (
	ParentNone      = 0x70000000
	ParentExtraMask = 0x80000000
	ParentLastMask  = 0x7fffffff

	GenerationOverflow = 0x80000000
)
