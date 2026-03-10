package sideband64k

// FrameType identifies the kind of decoded sideband frame.
type FrameType uint8

const (
	// FrameData carries primary payload bytes from band 1.
	FrameData FrameType = iota
	// FrameProgress carries progress bytes from band 2.
	FrameProgress
	// FrameError carries fatal error bytes from band 3.
	FrameError
	// FrameFlush is pkt-line control frame 0000.
	FrameFlush
	// FrameDelim is pkt-line control frame 0001.
	FrameDelim
	// FrameResponseEnd is pkt-line control frame 0002.
	FrameResponseEnd
)
