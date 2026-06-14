// Command idx-bloom reads a Git pack index
// and writes an IDBL Bloom filter over its object IDs to stdout.
//
// With an index filename argument the index is read from that file;
// with no argument it is read from stdin.
// A pack index does not record its object format,
// so the format must be given with -format.
package main
