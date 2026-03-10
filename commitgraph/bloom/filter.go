package bloom

// Filter represents a changed-paths Bloom filter associated with a commit.
//
// The filter encodes which paths changed between a commit and its first
// parent. Paths are expected to be in Git's slash-separated form and
// are queried using a path and its prefixes (e.g. "a/b/c", "a/b", "a").
type Filter struct {
	Data []byte

	HashVersion    uint32
	NumHashes      uint32
	BitsPerEntry   uint32
	MaxChangePaths uint32
}

// NewFilter constructs one query-ready bloom filter from raw data/settings.
func NewFilter(data []byte, settings Settings) *Filter {
	out := &Filter{
		Data:           data,
		HashVersion:    settings.HashVersion,
		NumHashes:      settings.NumHashes,
		BitsPerEntry:   settings.BitsPerEntry,
		MaxChangePaths: settings.MaxChangePaths,
	}

	return out
}
