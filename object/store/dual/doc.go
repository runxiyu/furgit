// Package dual provides one logical object store backed by separate object-wise
// and pack-wise stores.
//
// Dual composes a store that handles individual object writes with a store that
// handles pack-wise writes, while exposing one mixed reader over both.
// Coordinated quarantine operations span both stores, but quarantine promotion
// is non-atomic.
package dual
