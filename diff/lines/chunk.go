package lines

// Chunk represents a contiguous region of lines categorized
// as unchanged, deleted, or added.
type Chunk struct {
	Kind ChunkKind
	Data []byte
}

// ChunkKind enumerates the type of diff chunk.
type ChunkKind int

const (
	// ChunkKindUnchanged represents an unchanged diff chunk.
	ChunkKindUnchanged ChunkKind = iota
	// ChunkKindDeleted represents a deleted diff chunk.
	ChunkKindDeleted
	// ChunkKindAdded represents an added diff chunk.
	ChunkKindAdded
)
