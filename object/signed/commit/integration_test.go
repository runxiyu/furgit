package signedcommit_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	signedcommit "lindenii.org/go/furgit/object/signed/commit"
)

func setupSSHSignedCommit(
	t *testing.T,
	algo objectid.Algorithm,
) (payload []byte, allowedSignersPath string, signaturePath string) {
	t.Helper()

	testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})

	signDir := t.TempDir()

	signRoot, err := os.OpenRoot(signDir)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", signDir, err)
	}

	t.Cleanup(func() { _ = signRoot.Close() })

	privateKeyPath := filepath.Join(signDir, "signing_key")
	allowedSignersPath = filepath.Join(signDir, "allowed_signers")
	signaturePath = filepath.Join(signDir, "commit.sig")

	cmd := exec.Command( //nolint:noctx
		"ssh-keygen",
		"-q",
		"-t", "ed25519",
		"-N", "",
		"-C", "runxiyu@umich.edu",
		"-f", privateKeyPath,
	) //#nosec G204

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ssh-keygen generate failed: %v\n%s", err, out)
	}

	publicKey, err := signRoot.ReadFile("signing_key.pub")
	if err != nil {
		t.Fatalf("ReadFile(signing_key.pub): %v", err)
	}

	err = signRoot.WriteFile(
		"allowed_signers",
		append([]byte("runxiyu@umich.edu "), publicKey...),
		0o600,
	)
	if err != nil {
		t.Fatalf("WriteFile(allowed_signers): %v", err)
	}

	testRepo.Run(t, "config", "gpg.format", "ssh")
	testRepo.Run(t, "config", "user.signingkey", privateKeyPath)

	testRepo.WriteFile(t, "file.txt", []byte("signed\n"), 0o644)
	testRepo.Run(t, "add", "file.txt")
	testRepo.Run(t, "commit", "-S", "-m", "signed commit")

	commitID := testRepo.RevParse(t, "HEAD^{commit}")
	body := testRepo.CatFile(t, "commit", commitID)

	commit, err := signedcommit.Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	signature, ok := commit.AppendSignature(nil, algo)
	if !ok {
		t.Fatalf("missing %s signature", algo)
	}

	err = signRoot.WriteFile("commit.sig", signature, 0o600)
	if err != nil {
		t.Fatalf("WriteFile(commit.sig): %v", err)
	}

	return commit.AppendPayload(nil), allowedSignersPath, signaturePath
}

func TestSSHSignedCommitIntegration(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		payload, allowedSignersPath, signaturePath := setupSSHSignedCommit(t, algo)

		cmd := exec.Command( //nolint:noctx
			"ssh-keygen",
			"-Y", "verify",
			"-n", "git",
			"-f", allowedSignersPath,
			"-I", "runxiyu@umich.edu",
			"-s", signaturePath,
		) //#nosec G204
		cmd.Stdin = bytes.NewReader(payload)

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ssh-keygen verify failed: %v\n%s", err, out)
		}
	})
}

func TestSSHSignedCommitIntegrationRejectsTamperedPayload(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		payload, allowedSignersPath, signaturePath := setupSSHSignedCommit(t, algo)
		payload = append([]byte(nil), payload...)
		payload[len(payload)-2] ^= 1

		cmd := exec.Command( //nolint:noctx
			"ssh-keygen",
			"-Y", "verify",
			"-n", "git",
			"-f", allowedSignersPath,
			"-I", "runxiyu@umich.edu",
			"-s", signaturePath,
		) //#nosec G204
		cmd.Stdin = bytes.NewReader(payload)

		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("ssh-keygen verify unexpectedly succeeded:\n%s", out)
		}
	})
}
