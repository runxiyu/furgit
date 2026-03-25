package sideband64k

// Frame is one decoded side-band-64k frame.
//
// For FrameData, FrameProgress, and FrameError, Payload holds frame bytes and
// may be empty.
//
// For control frames, Payload is nil.
type Frame struct {
	Type    FrameType
	Payload []byte
}
