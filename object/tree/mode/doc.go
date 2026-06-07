// Package mode provides canonical Git tree entry modes.
//
// A [Mode] is the octal file mode stored in a tree entry,
// such as a regular file, executable, symlink, directory, or gitlink.
// It rejects malformed, unsupported, and zero-padded encodings.
package mode
