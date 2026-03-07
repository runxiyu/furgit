package receivepack_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"codeberg.org/lindenii/furgit/format/pktline"
	"codeberg.org/lindenii/furgit/format/sideband64k"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	receivepack "codeberg.org/lindenii/furgit/receivepack"
	receivepackhooks "codeberg.org/lindenii/furgit/receivepack/hooks"
)

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

func TestReceivePackHookProgressUsesSideBand64K(t *testing.T) {
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
			commitID.String() + " " + objectid.Zero(algo).String() + " refs/heads/main\x00report-status side-band-64k atomic delete-refs object-format=" + algo.String() + "\n",
		))
		input.WriteString("0000")

		err := receivepack.ReceivePack(context.Background(), &output, strings.NewReader(input.String()), receivepack.Options{
			Algorithm:       algo,
			Refs:            repo.Refs(),
			ExistingObjects: repo.Objects(),
			Hook: func(ctx context.Context, req receivepack.HookRequest) ([]receivepack.UpdateDecision, error) {
				_, err := io.WriteString(req.IO.Progress, "hook says hello\n")
				if err != nil {
					return nil, err
				}

				return []receivepack.UpdateDecision{{Accept: true}}, nil
			},
		})
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		_, sidebandWire, ok := strings.Cut(output.String(), "0000")
		if !ok {
			t.Fatalf("output missing advertisement flush: %q", output.String())
		}

		dec := sideband64k.NewDecoder(strings.NewReader(sidebandWire), sideband64k.ReadOptions{})

		frame, err := dec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(progress): %v", err)
		}

		if frame.Type != sideband64k.FrameProgress || string(frame.Payload) != "hook says hello\n" {
			t.Fatalf("first frame = %#v", frame)
		}

		frame, err = dec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(unpack): %v", err)
		}

		if frame.Type != sideband64k.FrameData {
			t.Fatalf("second frame = %#v", frame)
		}

		statusDec := pktline.NewDecoder(strings.NewReader(string(frame.Payload)), pktline.ReadOptions{})

		statusFrame, err := statusDec.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame(status unpack): %v", err)
		}

		if statusFrame.Type != pktline.PacketData || string(statusFrame.Payload) != "unpack ok\n" {
			t.Fatalf("status frame = %#v", statusFrame)
		}
	})
}

func TestReceivePackPredefinedRejectForcePushHookRejectsNonFastForward(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, treeID := testRepo.MakeSingleFileTree(t, "base.txt", []byte("base\n"))
		baseID := testRepo.CommitTree(t, treeID, "base")
		currentID := testRepo.CommitTree(t, treeID, "current", baseID)
		forcedID := testRepo.CommitTree(t, treeID, "forced", baseID)
		testRepo.UpdateRef(t, "refs/heads/main", currentID)

		repo := testRepo.OpenRepository(t)
		objectsRoot := testRepo.OpenObjectsRoot(t)
		packStream := testRepo.PackObjectsReader(t, []string{forcedID.String(), "^" + currentID.String()}, false)
		t.Cleanup(func() {
			_ = packStream.Close()
		})

		var (
			input  strings.Builder
			output bufferWriteFlusher
		)

		input.WriteString(pktlineData(
			currentID.String() + " " + forcedID.String() + " refs/heads/main\x00report-status atomic object-format=" + algo.String() + "\n",
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
				Hook:            receivepackhooks.RejectForcePush(),
			},
		)
		if err != nil {
			t.Fatalf("ReceivePack: %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ng refs/heads/main non-fast-forward\n") {
			t.Fatalf("unexpected receive-pack output %q", got)
		}

		resolved, err := repo.Refs().ResolveFully("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveFully(main): %v", err)
		}

		if resolved.ID != currentID {
			t.Fatalf("refs/heads/main = %s, want %s", resolved.ID, currentID)
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

func TestReceivePackGitPushCreatesBranch(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Parallel()

		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := sender.MakeCommit(t, "pushed commit")
		sender.UpdateRef(t, "refs/heads/main", commitID)

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		repo := receiver.OpenRepository(t)
		objectsRoot := receiver.OpenObjectsRoot(t)

		stdout, stderr, clientErr, serverErr := runGitPushFD(
			t,
			sender,
			receivepack.Options{
				Algorithm:       algo,
				Refs:            repo.Refs(),
				ExistingObjects: repo.Objects(),
				ObjectsRoot:     objectsRoot,
			},
			"push", "--porcelain", "fd::3,4/test", "refs/heads/main:refs/heads/main",
		)
		if clientErr != nil {
			t.Fatalf("git push failed: %v\nstdout=%s\nstderr=%s", clientErr, stdout, stderr)
		}

		if serverErr != nil {
			t.Fatalf("ReceivePack: %v", serverErr)
		}

		resolved, err := receiver.OpenRepository(t).Refs().ResolveFully("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveFully(main): %v", err)
		}

		if resolved.ID != commitID {
			t.Fatalf("refs/heads/main = %s, want %s", resolved.ID, commitID)
		}
	})
}

func TestReceivePackGitPushAtomicDelete(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Parallel()

		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := receiver.MakeCommit(t, "base")
		receiver.UpdateRef(t, "refs/heads/main", commitID)

		repo := receiver.OpenRepository(t)

		stdout, stderr, clientErr, serverErr := runGitPushFD(
			t,
			sender,
			receivepack.Options{
				Algorithm:       algo,
				Refs:            repo.Refs(),
				ExistingObjects: repo.Objects(),
			},
			"push", "--porcelain", "--atomic", "fd::3,4/test", ":refs/heads/main",
		)
		if clientErr != nil {
			t.Fatalf("git push failed: %v\nstdout=%s\nstderr=%s", clientErr, stdout, stderr)
		}

		if serverErr != nil {
			t.Fatalf("ReceivePack: %v", serverErr)
		}

		_, err := receiver.OpenRepository(t).Refs().Resolve("refs/heads/main")
		if err == nil {
			t.Fatal("refs/heads/main still exists after delete push")
		}
	})
}

func TestReceivePackGitPushRejectsForcedUpdateViaHook(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Parallel()

		sender := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		blobID, treeID := sender.MakeSingleFileTree(t, "base.txt", []byte("base\n"))
		baseID := sender.CommitTree(t, treeID, "base")
		currentID := sender.CommitTree(t, treeID, "current", baseID)
		forcedID := sender.CommitTree(t, treeID, "forced", baseID)
		sender.UpdateRef(t, "refs/heads/main", forcedID)

		receiver := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		receiver.HashObject(t, "blob", sender.RunBytes(t, "cat-file", "blob", blobID.String()))
		receiver.HashObject(t, "tree", sender.RunBytes(t, "cat-file", "tree", treeID.String()))
		receiver.HashObject(t, "commit", sender.RunBytes(t, "cat-file", "commit", baseID.String()))
		receiver.HashObject(t, "commit", sender.RunBytes(t, "cat-file", "commit", currentID.String()))
		receiver.UpdateRef(t, "refs/heads/main", currentID)

		repo := receiver.OpenRepository(t)
		objectsRoot := receiver.OpenObjectsRoot(t)

		stdout, stderr, clientErr, serverErr := runGitPushFD(
			t,
			sender,
			receivepack.Options{
				Algorithm:       algo,
				Refs:            repo.Refs(),
				ExistingObjects: repo.Objects(),
				ObjectsRoot:     objectsRoot,
				Hook:            receivepackhooks.RejectForcePush(),
			},
			"push", "--porcelain", "--force", "fd::3,4/test", "refs/heads/main:refs/heads/main",
		)
		if clientErr == nil {
			t.Fatalf("git push unexpectedly succeeded\nstdout=%s\nstderr=%s", stdout, stderr)
		}

		if serverErr != nil {
			t.Fatalf("ReceivePack: %v", serverErr)
		}

		if !strings.Contains(stdout, "non-fast-forward") && !strings.Contains(stderr, "non-fast-forward") {
			t.Fatalf("git push output missing non-fast-forward message\nstdout=%s\nstderr=%s", stdout, stderr)
		}

		resolved, err := receiver.OpenRepository(t).Refs().ResolveFully("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveFully(main): %v", err)
		}

		if resolved.ID != currentID {
			t.Fatalf("refs/heads/main = %s, want %s", resolved.ID, currentID)
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

type fileWriteFlusher struct {
	*os.File
}

func (fileWriteFlusher) Flush() error {
	return nil
}

func runGitPushFD(
	tb testing.TB,
	sender *testgit.TestRepo,
	opts receivepack.Options,
	gitArgs ...string,
) (stdout string, stderr string, clientErr error, serverErr error) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverRead, clientWrite, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe(serverRead/clientWrite): %v", err)
	}

	clientRead, serverWrite, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe(clientRead/serverWrite): %v", err)
	}

	tb.Cleanup(func() {
		_ = serverRead.Close()
		_ = clientWrite.Close()
		_ = clientRead.Close()
		_ = serverWrite.Close()
	})

	go func() {
		<-ctx.Done()

		_ = serverRead.Close()
		_ = clientWrite.Close()
		_ = clientRead.Close()
		_ = serverWrite.Close()
	}()

	serverErrCh := make(chan error, 1)

	go func() {
		defer func() {
			_ = serverRead.Close()
			_ = serverWrite.Close()
		}()

		serverErrCh <- receivepack.ReceivePack(
			ctx,
			fileWriteFlusher{serverWrite},
			serverRead,
			opts,
		)
	}()

	stdoutBytes, stderrBytes, clientErr := sender.RunWithExtraFilesEnvContextE(
		tb,
		ctx,
		nil,
		[]*os.File{clientRead, clientWrite},
		gitArgs...,
	)
	_ = clientRead.Close()
	_ = clientWrite.Close()

	serverErr = <-serverErrCh

	if ctx.Err() != nil {
		tb.Fatalf(
			"git push fd:: timed out\nstdout=%s\nstderr=%s\nclientErr=%v\nserverErr=%v",
			stdoutBytes,
			stderrBytes,
			clientErr,
			serverErr,
		)
	}

	return string(stdoutBytes), string(stderrBytes), clientErr, serverErr
}
