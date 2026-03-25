package receivepack

import (
	"fmt"
	"slices"
	"strings"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Capabilities describes one receive-pack capability set.
type Capabilities struct {
	ReportStatus   bool
	ReportStatusV2 bool
	DeleteRefs     bool
	SideBand64K    bool
	Quiet          bool
	Atomic         bool
	OfsDelta       bool
	PushOptions    bool
	PushCertNonce  string
	ObjectFormat   objectid.Algorithm
	SessionID      string
	Agent          string
}

// Normalize returns one normalized copy of caps.
func (caps Capabilities) Normalize(defaultAlgorithm objectid.Algorithm) Capabilities {
	if caps.ObjectFormat == objectid.AlgorithmUnknown {
		caps.ObjectFormat = defaultAlgorithm
	}

	return caps
}

// Tokens returns capabilities in Git advertisement order.
func (caps Capabilities) Tokens(defaultAlgorithm objectid.Algorithm) []string {
	caps = caps.Normalize(defaultAlgorithm)

	tokens := make([]string, 0, 11)
	if caps.ReportStatus {
		tokens = append(tokens, "report-status")
	}

	if caps.ReportStatusV2 {
		tokens = append(tokens, "report-status-v2")
	}

	if caps.DeleteRefs {
		tokens = append(tokens, "delete-refs")
	}

	if caps.SideBand64K {
		tokens = append(tokens, "side-band-64k")
	}

	if caps.Quiet {
		tokens = append(tokens, "quiet")
	}

	if caps.Atomic {
		tokens = append(tokens, "atomic")
	}

	if caps.OfsDelta {
		tokens = append(tokens, "ofs-delta")
	}

	if caps.PushCertNonce != "" {
		tokens = append(tokens, "push-cert="+caps.PushCertNonce)
	}

	if caps.PushOptions {
		tokens = append(tokens, "push-options")
	}

	if caps.SessionID != "" {
		tokens = append(tokens, "session-id="+caps.SessionID)
	}

	if caps.ObjectFormat != objectid.AlgorithmUnknown {
		tokens = append(tokens, "object-format="+caps.ObjectFormat.String())
	}

	if caps.Agent != "" {
		tokens = append(tokens, "agent="+caps.Agent)
	}

	return tokens
}

func (caps Capabilities) supportsToken(token string, defaultAlgorithm objectid.Algorithm) bool {
	name, value, _ := strings.Cut(token, "=")

	switch name {
	case "report-status":
		return caps.ReportStatus && value == ""
	case "report-status-v2":
		return caps.ReportStatusV2 && value == ""
	case "delete-refs":
		return caps.DeleteRefs && value == ""
	case "side-band-64k":
		return caps.SideBand64K && value == ""
	case "quiet":
		return caps.Quiet && value == ""
	case "atomic":
		return caps.Atomic && value == ""
	case "ofs-delta":
		return caps.OfsDelta && value == ""
	case "push-options":
		return caps.PushOptions && value == ""
	case "push-cert":
		return caps.PushCertNonce != "" && value != ""
	case "object-format":
		if value == "" {
			return false
		}

		algo, ok := objectid.ParseAlgorithm(value)

		return ok && algo == caps.Normalize(defaultAlgorithm).ObjectFormat
	case "session-id":
		return caps.SessionID != "" && value != ""
	case "agent":
		return caps.Agent != "" && value != ""
	default:
		return false
	}
}

func parseCapabilityList(s string) ([]string, error) {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil, nil
	}

	tokens := strings.Fields(s)
	if slices.Contains(tokens, "") {
		return nil, &ProtocolError{Reason: "empty capability token"}
	}

	return tokens, nil
}

func parseRequestedCapabilities(
	tokens []string,
	supported Capabilities,
	defaultAlgorithm objectid.Algorithm,
) (Capabilities, error) {
	var requested Capabilities

	requested.ObjectFormat = defaultAlgorithm

	for _, token := range tokens {
		if !supported.supportsToken(token, defaultAlgorithm) {
			return Capabilities{}, &ProtocolError{
				Reason: fmt.Sprintf("unsupported capability %q", token),
			}
		}

		name, value, _ := strings.Cut(token, "=")
		switch name {
		case "report-status":
			requested.ReportStatus = true
		case "report-status-v2":
			requested.ReportStatusV2 = true
		case "delete-refs":
			requested.DeleteRefs = true
		case "side-band-64k":
			requested.SideBand64K = true
		case "quiet":
			requested.Quiet = true
		case "atomic":
			requested.Atomic = true
		case "ofs-delta":
			requested.OfsDelta = true
		case "push-options":
			requested.PushOptions = true
		case "push-cert":
			requested.PushCertNonce = value
		case "object-format":
			algo, _ := objectid.ParseAlgorithm(value)
			requested.ObjectFormat = algo
		case "session-id":
			requested.SessionID = value
		case "agent":
			requested.Agent = value
		}
	}

	return requested, nil
}
