package receivepack_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	receivepack "codeberg.org/lindenii/furgit/receivepack"
)

// TODO: actually test with send-pack

func TestReceivePackDeleteOnlyReportsNotImplemented(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, &strings.Builder{}, strings.NewReader(input.String()), receivepack.Options{
			GitProtocol:     "",
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ng refs/heads/main ref updates not implemented yet\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}
	})
}

func TestReceivePackAdvertisesResolvedHEAD(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, &strings.Builder{}, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()

		want := commitID.String() + " HEAD"
		if !strings.Contains(got, want) {
			t.Fatalf("HEAD advertisement missing %q in %q", want, got)
		}
	})
}

func TestReceivePackVersion2FallsBackToV0(t *testing.T) {
	t.Parallel()

	testReceivePackProtocolFallback(t, "version=2")
}

func TestReceivePackHighestRequestedVersionFallsBackToV0ForV2(t *testing.T) {
	t.Parallel()

	testReceivePackProtocolFallback(t, "version=1:version=2")
}

func TestReceivePackWithoutReportStatusWritesNoStatusPayload(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, &strings.Builder{}, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if strings.Contains(got, "unpack ") || strings.Contains(got, "ng refs/heads/main ") || strings.Contains(got, "ok refs/heads/main\n") {
			t.Fatalf("unexpected status payload %q", got)
		}
	})
}

func testReceivePackProtocolFallback(t *testing.T, gitProtocol string) {
	t.Helper()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, &strings.Builder{}, strings.NewReader(input.String()), receivepack.Options{
			GitProtocol:     gitProtocol,
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		if strings.HasPrefix(output.String(), pktlineData("version 1\n")) {
			t.Fatalf("receive-pack output started with protocol v1 preface for %q: %q", gitProtocol, output.String())
		}
	})
}

func TestReceivePackPackRequestWithoutObjectsRootReportsNotConfigured(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			commitID.String() + " " + commitID.String() + " refs/heads/main\x00report-status object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, &strings.Builder{}, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "unpack objects root not configured\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}
	})
}

type bufferWriteFlusher struct {
	strings.Builder
}

func (bufferWriteFlusher) Flush() error {
	return nil
}

func pktlineData(payload string) string {
	return fmt.Sprintf("%04x%s", len(payload)+4, payload)
}
