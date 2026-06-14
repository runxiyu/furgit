// Command explain-pack reads a Git packfile and writes a
// human-readable explanation to stdout.
//
// With a pack filename argument
// the pack is mmap'd
// and a sibling .idx is used when present;
// with no argument the pack is read from stdin.
// A packfile does not record its object format,
// so the format must be given with -format.
package main
