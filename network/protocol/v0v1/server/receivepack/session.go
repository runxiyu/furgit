package receivepack

import (
	"fmt"
	"io"
	"strings"

	"codeberg.org/lindenii/furgit/common/iowrap"
	common "codeberg.org/lindenii/furgit/network/protocol/v0v1/server"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Session is one stateful server-side receive-pack protocol session.
type Session struct {
	base       *common.Session
	supported  Capabilities
	negotiated Capabilities
}

// NewSession creates one receive-pack session over one common server session.
func NewSession(base *common.Session, supported Capabilities) *Session {
	return &Session{
		base:      base,
		supported: supported,
	}
}

// AdvertiseRefs writes one receive-pack ref advertisement.
func (session *Session) AdvertiseRefs(ad common.Advertisement) error {
	return session.base.AdvertiseRefs(ad, session.supported.Tokens(session.base.Algorithm()))
}

// ReadRequest reads one receive-pack request through optional push-options.
func (session *Session) ReadRequest() (*Request, error) {
	req := &Request{}

	var sawCommands bool

	for {
		frame, err := session.base.ReadFrame()
		if err != nil {
			return nil, err
		}

		switch frame.Type {
		case common.FrameFlush:
			goto afterCommands
		case common.FrameData:
		case common.FrameDelim, common.FrameResponseEnd:
			return nil, &ProtocolError{Reason: fmt.Sprintf("unexpected packet type %v", frame.Type)}
		}

		payload := string(frame.Payload)
		if strings.HasPrefix(payload, "shallow ") {
			line := trimOneLF(payload)

			shallowID, err := parseObjectID(session.base.Algorithm(), line[len("shallow "):])
			if err != nil {
				return nil, err
			}

			req.Shallow = append(req.Shallow, shallowID)

			continue
		}

		if strings.HasPrefix(payload, "push-cert\x00") {
			if sawCommands {
				return nil, &ProtocolError{Reason: "got both push certificate and unsigned commands"}
			}

			capabilityTokens, err := parseCapabilityList(payload[len("push-cert\x00"):])
			if err != nil {
				return nil, err
			}

			requested, err := parseRequestedCapabilities(
				capabilityTokens,
				session.supported,
				session.base.Algorithm(),
			)
			if err != nil {
				return nil, err
			}

			req.Capabilities = requested

			cert, err := session.readPushCertificate()
			if err != nil {
				return nil, err
			}

			req.PushCert = cert
			req.Commands = append(req.Commands, cert.Commands...)
			sawCommands = true

			continue
		}

		line := trimOneLF(payload)
		if !sawCommands && strings.Contains(line, "\x00") {
			commandPart, capPart, _ := strings.Cut(line, "\x00")

			capabilityTokens, err := parseCapabilityList(capPart)
			if err != nil {
				return nil, err
			}

			requested, err := parseRequestedCapabilities(
				capabilityTokens,
				session.supported,
				session.base.Algorithm(),
			)
			if err != nil {
				return nil, err
			}

			req.Capabilities = requested
			line = commandPart
		}

		cmd, err := parseCommand(session.base.Algorithm(), line)
		if err != nil {
			return nil, err
		}

		req.Commands = append(req.Commands, cmd)
		sawCommands = true
	}

afterCommands:
	if req.Capabilities.PushOptions {
		for {
			frame, err := session.base.ReadFrame()
			if err != nil {
				return nil, err
			}

			switch frame.Type {
			case common.FrameFlush:
				goto afterPushOptions
			case common.FrameData:
				req.PushOptions = append(req.PushOptions, trimOneLF(string(frame.Payload)))
			case common.FrameDelim, common.FrameResponseEnd:
				return nil, &ProtocolError{Reason: fmt.Sprintf("unexpected packet type %v", frame.Type)}
			}
		}
	}

afterPushOptions:
	req.DeleteOnly = deleteOnly(req.Commands)

	req.PackExpected = len(req.Commands) > 0 && !req.DeleteOnly

	session.negotiated = req.Capabilities

	if req.Capabilities.SideBand64K {
		session.base.EnableSideBand64K()
	}

	return req, nil
}

// WriteProgress writes one progress packet.
func (session *Session) WriteProgress(p []byte) error {
	return session.base.WriteProgress(p)
}

// ProgressWriter returns one chunking writer for sideband progress output.
//
// When side-band-64k was not negotiated, writes are discarded.
func (session *Session) ProgressWriter() iowrap.WriteFlusher {
	return session.base.ProgressWriter()
}

// WriteError writes one fatal error packet.
func (session *Session) WriteError(p []byte) error {
	return session.base.WriteError(p)
}

// ErrorWriter returns one chunking writer for sideband error output.
//
// When side-band-64k was not negotiated, writes are discarded.
func (session *Session) ErrorWriter() io.Writer {
	return session.base.ErrorWriter()
}

func trimOneLF(s string) string {
	return strings.TrimSuffix(s, "\n")
}

func parseObjectID(algo objectid.Algorithm, s string) (objectid.ObjectID, error) {
	id, err := objectid.ParseHex(algo, s)
	if err != nil {
		return objectid.ObjectID{}, &ProtocolError{
			Reason: fmt.Sprintf("invalid object id %q", s),
		}
	}

	return id, nil
}

func commandIsDelete(cmd Command) bool {
	return cmd.NewID == objectid.Zero(cmd.NewID.Algorithm())
}

func deleteOnly(commands []Command) bool {
	if len(commands) == 0 {
		return false
	}

	for _, cmd := range commands {
		if !commandIsDelete(cmd) {
			return false
		}
	}

	return true
}

func parseCommand(algo objectid.Algorithm, line string) (Command, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return Command{}, &ProtocolError{Reason: fmt.Sprintf("malformed command %q", line)}
	}

	oldID, err := parseObjectID(algo, fields[0])
	if err != nil {
		return Command{}, err
	}

	newID, err := parseObjectID(algo, fields[1])
	if err != nil {
		return Command{}, err
	}

	return Command{OldID: oldID, NewID: newID, Name: fields[2]}, nil
}

func (session *Session) readPushCertificate() (*PushCertificate, error) {
	cert := &PushCertificate{}
	inCommands := false
	inSignature := false

	for {
		frame, err := session.base.ReadFrame()
		if err != nil {
			return nil, err
		}

		switch frame.Type {
		case common.FrameFlush:
			return nil, &ProtocolError{Reason: "unexpected flush inside push certificate"}
		case common.FrameData:
		case common.FrameDelim, common.FrameResponseEnd:
			return nil, &ProtocolError{Reason: fmt.Sprintf("unexpected packet type %v", frame.Type)}
		}

		line := string(frame.Payload)
		if line == "push-cert-end\n" {
			return cert, nil
		}

		if !inCommands {
			if line == "\n" {
				inCommands = true

				continue
			}

			trimmed := trimOneLF(line)
			cert.HeaderLines = append(cert.HeaderLines, trimmed)

			if strings.HasPrefix(trimmed, "push-option ") {
				cert.EmbeddedOption = append(cert.EmbeddedOption, trimmed[len("push-option "):])
			}

			continue
		}

		if !inSignature {
			trimmed := trimOneLF(line)

			cmd, err := parseCommand(session.base.Algorithm(), trimmed)
			if err == nil {
				cert.Commands = append(cert.Commands, cmd)

				continue
			}

			inSignature = true
		}

		cert.SignatureLines = append(cert.SignatureLines, trimOneLF(line))
	}
}
