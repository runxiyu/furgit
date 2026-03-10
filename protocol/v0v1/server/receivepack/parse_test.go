package receivepack_test

import (
	"errors"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/protocol/pktline"
	common "codeberg.org/lindenii/furgit/protocol/v0v1/server"
	receivepack "codeberg.org/lindenii/furgit/protocol/v0v1/server/receivepack"
)

func TestReadRequestParsesCommandsAndPushOptions(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		oldZero := objectid.Zero(algo).String()
		oneID := mustHexID(t, algo, "1")

		var wire bufferWriteFlusher

		enc := pktline.NewEncoder(&wire)

		err := enc.WriteData([]byte(
			oldZero + " " + oneID.String() + " refs/heads/main\x00report-status push-options object-format=" + algo.String() + "\n",
		))
		if err != nil {
			t.Fatalf("WriteData(first): %v", err)
		}

		err = enc.WriteData([]byte(
			oneID.String() + " " + oldZero + " refs/heads/old\n",
		))
		if err != nil {
			t.Fatalf("WriteData(second): %v", err)
		}

		err = enc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush(commands): %v", err)
		}

		err = enc.WriteData([]byte("ci.skip\n"))
		if err != nil {
			t.Fatalf("WriteData(push-option): %v", err)
		}

		err = enc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush(push-options): %v", err)
		}

		base := common.NewSession(strings.NewReader(wire.String()), &bufferWriteFlusher{}, common.Options{
			Algorithm: algo,
		})
		session := receivepack.NewSession(base, receivepack.Capabilities{
			ReportStatus: true,
			PushOptions:  true,
			ObjectFormat: algo,
		})

		req, err := session.ReadRequest()
		if err != nil {
			t.Fatalf("ReadRequest: %v", err)
		}

		if len(req.Commands) != 2 {
			t.Fatalf("len(req.Commands) = %d, want 2", len(req.Commands))
		}

		if !req.Capabilities.ReportStatus || !req.Capabilities.PushOptions {
			t.Fatalf("capabilities = %#v", req.Capabilities)
		}

		if len(req.PushOptions) != 1 || req.PushOptions[0] != "ci.skip" {
			t.Fatalf("push options = %#v", req.PushOptions)
		}

		if !req.PackExpected {
			t.Fatalf("PackExpected = false, want true")
		}

		if req.DeleteOnly {
			t.Fatalf("DeleteOnly = true, want false")
		}
	})
}

func TestReadRequestDeleteOnlyDoesNotExpectPack(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		oneID := mustHexID(t, algo, "1")

		var wire bufferWriteFlusher

		enc := pktline.NewEncoder(&wire)

		err := enc.WriteData([]byte(
			oneID.String() + " " + objectid.Zero(algo).String() + " refs/heads/old\x00delete-refs object-format=" + algo.String() + "\n",
		))
		if err != nil {
			t.Fatalf("WriteData: %v", err)
		}

		err = enc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush: %v", err)
		}

		base := common.NewSession(strings.NewReader(wire.String()), &bufferWriteFlusher{}, common.Options{
			Algorithm: algo,
		})
		session := receivepack.NewSession(base, receivepack.Capabilities{
			DeleteRefs:   true,
			ObjectFormat: algo,
		})

		req, err := session.ReadRequest()
		if err != nil {
			t.Fatalf("ReadRequest: %v", err)
		}

		if req.PackExpected {
			t.Fatalf("PackExpected = true, want false")
		}

		if !req.DeleteOnly {
			t.Fatalf("DeleteOnly = false, want true")
		}
	})
}

func TestReadRequestRejectsUnsupportedCapability(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		oneID := mustHexID(t, algo, "1")

		var wire bufferWriteFlusher

		enc := pktline.NewEncoder(&wire)

		err := enc.WriteData([]byte(
			objectid.Zero(algo).String() + " " + oneID.String() + " refs/heads/main\x00atomic object-format=" + algo.String() + "\n",
		))
		if err != nil {
			t.Fatalf("WriteData: %v", err)
		}

		err = enc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush: %v", err)
		}

		base := common.NewSession(strings.NewReader(wire.String()), &bufferWriteFlusher{}, common.Options{
			Algorithm: algo,
		})
		session := receivepack.NewSession(base, receivepack.Capabilities{ObjectFormat: algo})

		_, err = session.ReadRequest()
		if err == nil {
			t.Fatalf("ReadRequest error = nil, want error")
		}

		protocolErr, ok := errors.AsType[*receivepack.ProtocolError](err)
		if !ok {
			t.Fatalf("errors.AsType[*receivepack.ProtocolError](%T) = false", err)
		}

		if !strings.Contains(protocolErr.Reason, "unsupported capability") {
			t.Fatalf("ProtocolError.Reason = %q", protocolErr.Reason)
		}
	})
}

func TestReadRequestParsesPushCertificate(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		oneID := mustHexID(t, algo, "1")

		var wire bufferWriteFlusher

		enc := pktline.NewEncoder(&wire)

		err := enc.WriteData([]byte("push-cert\x00push-cert=nonce object-format=" + algo.String() + "\n"))
		if err != nil {
			t.Fatalf("WriteData(push-cert): %v", err)
		}

		lines := []string{
			"certificate version 0.1\n",
			"pusher Example <example@example.com>\n",
			"nonce nonce\n",
			"push-option ci.skip\n",
			"\n",
			objectid.Zero(algo).String() + " " + oneID.String() + " refs/heads/main\n",
			"-----BEGIN PGP SIGNATURE-----\n",
			"abcdef\n",
			"push-cert-end\n",
		}

		for _, line := range lines {
			err = enc.WriteData([]byte(line))
			if err != nil {
				t.Fatalf("WriteData(%q): %v", line, err)
			}
		}

		err = enc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush: %v", err)
		}

		base := common.NewSession(strings.NewReader(wire.String()), &bufferWriteFlusher{}, common.Options{
			Algorithm: algo,
		})
		session := receivepack.NewSession(base, receivepack.Capabilities{
			PushCertNonce: "server-nonce",
			ObjectFormat:  algo,
		})

		req, err := session.ReadRequest()
		if err != nil {
			t.Fatalf("ReadRequest: %v", err)
		}

		if req.PushCert == nil {
			t.Fatalf("PushCert = nil, want parsed certificate")
		}

		if len(req.Commands) != 1 {
			t.Fatalf("len(req.Commands) = %d, want 1", len(req.Commands))
		}

		if len(req.PushCert.EmbeddedOption) != 1 || req.PushCert.EmbeddedOption[0] != "ci.skip" {
			t.Fatalf("embedded options = %#v", req.PushCert.EmbeddedOption)
		}
	})
}
