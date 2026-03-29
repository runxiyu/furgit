// Package signedcommit extracts commit signing payloads and signatures from raw
// commit object bodies.
package signedcommit

// TODO: Consider whether we want to fully copy the bytes into here.
// The Append functions are a bit weird ergonomically.
