package receivepack

// ProtocolError reports one malformed or unsupported receive-pack protocol input.
type ProtocolError struct {
	Reason string
}

// Error returns the formatted error string.
func (err *ProtocolError) Error() string {
	return "protocol/v0v1/server/receivepack: protocol error: " + err.Reason
}
