package receivepack_test

import (
	"errors"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/format/pktline"
	"codeberg.org/lindenii/furgit/format/sideband64k"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	common "codeberg.org/lindenii/furgit/protocol/v0v1/server"
	receivepack "codeberg.org/lindenii/furgit/protocol/v0v1/server/receivepack"
)

func TestWriteReportStatusWritesClassicStatus(t *testing.T) {
	t.Parallel()

	var out bufferWriteFlusher

	base := common.NewSession(strings.NewReader(""), &out, common.Options{})
	session := receivepack.NewSession(base, receivepack.Capabilities{})

	err := session.WriteReportStatus(receivepack.ReportStatusResult{
		Commands: []receivepack.CommandResult{
			{Name: "refs/heads/main"},
			{Name: "refs/heads/dev", Error: "non-fast-forward"},
		},
	})
	if err != nil {
		t.Fatalf("WriteReportStatus: %v", err)
	}

	got := out.String()
	wantParts := []string{
		"unpack ok\n",
		"ok refs/heads/main\n",
		"ng refs/heads/dev non-fast-forward\n",
		"0000",
	}

	for _, part := range wantParts {
		if !strings.Contains(got, part) {
			t.Fatalf("report-status missing %q in %q", part, got)
		}
	}
}

func TestWriteReportStatusUsesSideBand64KWhenNegotiated(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		var requestWire bufferWriteFlusher

		requestEnc := pktline.NewEncoder(&requestWire)

		err := requestEnc.WriteData([]byte(
			objectid.Zero(algo).String() + " " + mustHexID(t, algo, "1").String() + " refs/heads/main\x00report-status side-band-64k object-format=" + algo.String() + "\n",
		))
		if err != nil {
			t.Fatalf("WriteData(request): %v", err)
		}

		err = requestEnc.WriteFlush()
		if err != nil {
			t.Fatalf("WriteFlush(request): %v", err)
		}

		var out bufferWriteFlusher

		base := common.NewSession(strings.NewReader(requestWire.String()), &out, common.Options{
			Algorithm: algo,
		})
		session := receivepack.NewSession(base, receivepack.Capabilities{
			ReportStatus: true,
			SideBand64K:  true,
			ObjectFormat: algo,
		})

		_, err = session.ReadRequest()
		if err != nil {
			t.Fatalf("ReadRequest: %v", err)
		}

		err = session.WriteReportStatus(receivepack.ReportStatusResult{
			Commands: []receivepack.CommandResult{
				{Name: "refs/heads/main"},
			},
		})
		if err != nil {
			t.Fatalf("WriteReportStatus: %v", err)
		}

		dec := sideband64k.NewDecoder(strings.NewReader(out.String()), sideband64k.ReadOptions{})

		frame, err := dec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(unpack): %v", err)
		}

		if frame.Type != sideband64k.FrameData || string(frame.Payload) != "unpack ok\n" {
			t.Fatalf("first frame = %#v", frame)
		}

		frame, err = dec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(ok): %v", err)
		}

		if frame.Type != sideband64k.FrameData || string(frame.Payload) != "ok refs/heads/main\n" {
			t.Fatalf("second frame = %#v", frame)
		}

		frame, err = dec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(flush): %v", err)
		}

		if frame.Type != sideband64k.FrameFlush {
			t.Fatalf("flush frame.Type = %v, want FrameFlush", frame.Type)
		}
	})
}

func TestWriteReportStatusV2WritesOptionLines(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		oldID := mustHexID(t, algo, "1")
		newID := mustHexID(t, algo, "2")

		var out bufferWriteFlusher

		base := common.NewSession(strings.NewReader(""), &out, common.Options{})
		session := receivepack.NewSession(base, receivepack.Capabilities{})

		err := session.WriteReportStatusV2(receivepack.ReportStatusResult{
			Commands: []receivepack.CommandResult{
				{
					Name:         "refs/pseudo/proc",
					RefName:      "refs/heads/main",
					OldID:        &oldID,
					NewID:        &newID,
					ForcedUpdate: true,
				},
				{Name: "refs/heads/dev", Error: "rejected"},
			},
		})
		if err != nil {
			t.Fatalf("WriteReportStatusV2: %v", err)
		}

		got := out.String()
		wantParts := []string{
			"unpack ok\n",
			"ok refs/pseudo/proc\n",
			"option refname refs/heads/main\n",
			"option old-oid " + oldID.String() + "\n",
			"option new-oid " + newID.String() + "\n",
			"option forced-update\n",
			"ng refs/heads/dev rejected\n",
			"0000",
		}

		for _, part := range wantParts {
			if !strings.Contains(got, part) {
				t.Fatalf("report-status-v2 missing %q in %q", part, got)
			}
		}
	})
}

func TestWriteProgressRequiresSideBand64K(t *testing.T) {
	t.Parallel()

	base := common.NewSession(strings.NewReader(""), &bufferWriteFlusher{}, common.Options{})
	session := receivepack.NewSession(base, receivepack.Capabilities{})

	err := session.WriteProgress([]byte("progress\n"))
	if !errors.Is(err, common.ErrSideBandNotEnabled) {
		t.Fatalf("WriteProgress error = %v, want %v", err, common.ErrSideBandNotEnabled)
	}
}
