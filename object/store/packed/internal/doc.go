// Package internal encapsulates packed store implementation details.
//
// We have separate internal subpackages for ingest vs read and such,
// because these operations are so different that they almost share
// no code. This makes things clearer.
package internal
