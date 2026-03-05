package bloom

// Filter represents a changed-paths Bloom filter associated with a commit.
//
// The filter encodes which paths changed between a commit and its first
// parent. Paths are expected to be in Git's slash-separated form and
// are queried using a path and its prefixes (e.g. "a/b/c", "a/b", "a").
type Filter struct {
	Data    []byte
	Version uint32
}
