package server

// Version identifies the protocol version used on one v0/v1 server session.
type Version uint8

const (
	// Version0 is the original protocol framing with no leading version line.
	Version0 Version = iota
	// Version1 is protocol v1, which is v0 plus one leading "version 1\n"
	// pkt-line before ref advertisement.
	Version1
)
