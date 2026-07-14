// Package chain provides a read-only reference store
// that queries an ordered list of backends.
//
// The chain resolves each name against its backends in order,
// returning the first backend that has it,
// so earlier backends take precedence over later ones.
package chain
