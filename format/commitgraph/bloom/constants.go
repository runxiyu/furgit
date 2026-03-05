package bloom

const (
	// DataHeaderSize is the size of the BDAT header in commit-graph files.
	DataHeaderSize = 3 * 4
	// DefaultMaxChange matches Git's default max-changed-paths behavior.
	DefaultMaxChange = 512
)
