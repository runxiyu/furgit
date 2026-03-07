package server

// ProtocolError reports one malformed or unsupported protocol input.
type ProtocolError struct {
	Reason string
}

// Error returns the formatted error string.
func (err *ProtocolError) Error() string {
	return "protocol/v0v1/server: protocol error: " + err.Reason
}

// ErrUnexpectedPacket reports one unexpected pkt-line control packet.
var ErrUnexpectedPacket = &ProtocolError{Reason: "unexpected control packet"}

// ErrSideBandNotEnabled reports one attempt to write sideband frames without a
// negotiated side-band-64k session.
var ErrSideBandNotEnabled = &ProtocolError{Reason: "side-band-64k not enabled"}
