package receivepack_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	receivepack "codeberg.org/lindenii/furgit/receivepack"
)

// TODO: actually test with send-pack

func TestReceivePackDeleteOnlyAtomicDeleteSucceeds(t *testing.T) {
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
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status atomic delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			GitProtocol:     "",
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ok refs/heads/main\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		_, err = repo.Refs().Resolve("refs/heads/main")
		if err == nil {
			t.Fatal("refs/heads/main still exists after delete push")
		}
	})
}

func TestReceivePackDeleteOnlyNonAtomicAppliesIndependentDeletes(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		_, _, staleID := testRepo.MakeCommit(t, "stale")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/topic", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			staleID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/topic\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			GitProtocol:     "",
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ng refs/heads/main ") || !strings.Contains(got, "ok refs/heads/topic\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		_, err = repo.Refs().Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		_, err = repo.Refs().Resolve("refs/heads/topic")
		if err == nil {
			t.Fatal("refs/heads/topic still exists after successful delete")
		}
	})
}

func TestReceivePackDeleteOnlyAtomicFailureLeavesAllRefsUntouched(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		_, _, staleID := testRepo.MakeCommit(t, "stale")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/topic", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			staleID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status atomic delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/topic\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			GitProtocol:     "",
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ng refs/heads/main ") || !strings.Contains(got, "ng refs/heads/topic ") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		_, err = repo.Refs().Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		_, err = repo.Refs().Resolve("refs/heads/topic")
		if err != nil {
			t.Fatalf("Resolve(topic): %v", err)
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

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
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
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00delete-refs atomic object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
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
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status atomic delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
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

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
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

func TestReceivePackPackCreatePromotesObjectsAndUpdatesRef(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := sender.MakeCommit(t, "pushed commit")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		repo := receiver.OpenRepository(t)
		objectsRoot := receiver.OpenObjectsRoot(t)

		packStream := sender.PackObjectsReader(t, []string{commitID.String()}, false)
		t.Cleanup(func() {
			_ = packStream.Close()
		})

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			objectid.Zero(algo).String() + " " + commitID.String() + " refs/heads/main\x00report-status-v2 atomic object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(
			context.Background(),
			&output,
			io.MultiReader(strings.NewReader(input.String()), packStream),
			receivepack.Options{
				Algorithm:       algo,
				Refs:            repo.Refs(),
				ExistingObjects: repo.Objects(),
				ObjectsRoot:     objectsRoot,
			},
		)
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "unpack ok\n") || !strings.Contains(got, "ok refs/heads/main\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		reopened := receiver.OpenRepository(t)

		resolved, err := reopened.Refs().ResolveFully("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveFully(main): %v", err)
		}

		if resolved.ID != commitID {
			t.Fatalf("refs/heads/main = %s, want %s", resolved.ID, commitID)
		}

		if gotType := receiver.Run(t, "cat-file", "-t", commitID.String()); gotType != "commit" {
			t.Fatalf("cat-file -t = %q, want commit", gotType)
		}

		packs := receiver.Run(t, "count-objects", "-v")
		if !strings.Contains(packs, "packs: 1") {
			t.Fatalf("count-objects output missing promoted pack: %q", packs)
		}
	})
}

func TestReceivePackHookSeesQuarantinedObjectsAndCanRejectBeforePromotion(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := sender.MakeCommit(t, "pushed commit")

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		repo := receiver.OpenRepository(t)
		objectsRoot := receiver.OpenObjectsRoot(t)

		packStream := sender.PackObjectsReader(t, []string{commitID.String()}, false)
		t.Cleanup(func() {
			_ = packStream.Close()
		})

		var (
			input      strings.Builder
			output     bufferWriteFlusher
			hookCalled bool
		)

		input.WriteString(pktlineData(
			objectid.Zero(algo).String() + " " + commitID.String() + " refs/heads/main\x00report-status-v2 atomic object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(
			context.Background(),
			&output,
			io.MultiReader(strings.NewReader(input.String()), packStream),
			receivepack.Options{
				Algorithm:       algo,
				Refs:            repo.Refs(),
				ExistingObjects: repo.Objects(),
				ObjectsRoot:     objectsRoot,
				Hook: func(ctx context.Context, req receivepack.HookRequest) ([]receivepack.UpdateDecision, error) {
					hookCalled = true

					if len(req.Updates) != 1 || req.Updates[0].NewID != commitID {
						t.Fatalf("unexpected hook updates: %+v", req.Updates)
					}

					_, _, err := req.ExistingObjects.ReadHeader(commitID)
					if err == nil {
						t.Fatalf("existing objects unexpectedly contained quarantined commit %s", commitID)
					}

					_, _, err = req.QuarantinedObjects.ReadHeader(commitID)
					if err != nil {
						t.Fatalf("quarantined objects missing commit %s: %v", commitID, err)
					}

					return []receivepack.UpdateDecision{{
						Accept:  false,
						Message: "blocked by hook",
					}}, nil
				},
			},
		)
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		if !hookCalled {
			t.Fatal("hook was not called")
		}

		got := output.String()
		if !strings.Contains(got, "unpack ok\n") || !strings.Contains(got, "ng refs/heads/main blocked by hook\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		_, err = repo.Refs().Resolve("refs/heads/main")
		if err == nil {
			t.Fatal("refs/heads/main exists after hook rejection")
		}

		packs := receiver.Run(t, "count-objects", "-v")
		if !strings.Contains(packs, "packs: 0") {
			t.Fatalf("count-objects output shows unexpected promoted pack: %q", packs)
		}
	})
}

func TestReceivePackHookCanRejectSubsetOfNonAtomicDeleteOnlyPush(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/topic", commitID)

		repo := testRepo.OpenRepository(t)

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString(pktlineData(
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/topic\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
			Hook: func(ctx context.Context, req receivepack.HookRequest) ([]receivepack.UpdateDecision, error) {
				return []receivepack.UpdateDecision{
					{Accept: false, Message: "leave main alone"},
					{Accept: true},
				}, nil
			},
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ng refs/heads/main leave main alone\n") || !strings.Contains(got, "ok refs/heads/topic\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		_, err = repo.Refs().Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		_, err = repo.Refs().Resolve("refs/heads/topic")
		if err == nil {
			t.Fatal("refs/heads/topic still exists after successful delete")
		}
	})
}

func TestReceivePackReportStatusV2IncludesRefDetails(t *testing.T) {
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
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status-v2 atomic delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "option refname refs/heads/main\n") {
			t.Fatalf("missing option refname in %q", got)
		}

		if !strings.Contains(got, "option old-oid "+commitID.String()+"\n") {
			t.Fatalf("missing option old-oid in %q", got)
		}

		if !strings.Contains(got, "option new-oid "+objectid.Zero(algo).String()+"\n") {
			t.Fatalf("missing option new-oid in %q", got)
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
