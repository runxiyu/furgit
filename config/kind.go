package config

// Kind describes the presence and form of a config value.
type Kind uint8

const (
	// KindMissing means the queried key does not exist.
	KindMissing Kind = iota

	// KindValueless means the key exists but has no "= <value>" part.
	KindValueless

	// KindString means the key exists and has an explicit value (possibly "").
	KindString
)
