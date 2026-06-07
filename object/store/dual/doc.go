// Package dual provides one logical object store
// backed by separate object-wise and pack-wise stores.
//
// Dual composes a store that handles individual object writes
// with a store that handles pack-wise writes,
// while exposing one most-recently-used reader over both.
// Quarantines span the sides they cover.
package dual
