package commit_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signed/commit"
	"lindenii.org/go/furgit/object/typ"
)

const signerPrincipal = "signer@example.org"

func setupSSHSignedCommit(
	t *testing.T,
	objectFormat id.ObjectFormat,
) (payload []byte, allowedSignersPath string, signaturePath string) {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	signDir := t.TempDir()

	signRoot, err := os.OpenRoot(signDir)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q): %v", signDir, err)
	}

	t.Cleanup(func() { _ = signRoot.Close() })

	privateKeyPath := filepath.Join(signDir, "signing_key")
	allowedSignersPath = filepath.Join(signDir, "allowed_signers")
	signaturePath = filepath.Join(signDir, "commit.sig")

	cmd := exec.CommandContext(
		t.Context(),
		"ssh-keygen",
		"-q",
		"-t", "ed25519",
		"-N", "",
		"-C", signerPrincipal,
		"-f", privateKeyPath,
	) //#nosec G204

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ssh-keygen: %v\n%s", err, out)
	}

	publicKey, err := signRoot.ReadFile("signing_key.pub")
	if err != nil {
		t.Fatalf("ReadFile(signing_key.pub): %v", err)
	}

	err = signRoot.WriteFile(
		"allowed_signers",
		append([]byte(signerPrincipal+" "), publicKey...),
		0o600,
	)
	if err != nil {
		t.Fatalf("WriteFile(allowed_signers): %v", err)
	}

	err = repo.ConfigSet(t, "gpg.format", "ssh")
	if err != nil {
		t.Fatalf("ConfigSet(gpg.format): %v", err)
	}

	err = repo.ConfigSet(t, "user.signingkey", privateKeyPath)
	if err != nil {
		t.Fatalf("ConfigSet(user.signingkey): %v", err)
	}

	blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("signed\n"))
	if err != nil {
		t.Fatalf("HashObject(blob): %v", err)
	}

	treeID, err := repo.MkTree(t, []testgit.TreeEntry{
		{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file.txt"},
	})
	if err != nil {
		t.Fatalf("MkTree: %v", err)
	}

	commitID, err := repo.CommitTree(t, treeID, testgit.CommitTreeOptions{
		Message: "signed commit",
		Sign:    true,
	})
	if err != nil {
		t.Fatalf("CommitTree: %v", err)
	}

	body, err := repo.CatFile(t, typ.TypeCommit, commitID)
	if err != nil {
		t.Fatalf("CatFile: %v", err)
	}

	parsed, err := commit.Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	signature, ok := parsed.AppendSignature(nil, objectFormat)
	if !ok {
		t.Fatalf("missing %s signature", objectFormat)
	}

	err = signRoot.WriteFile("commit.sig", signature, 0o600)
	if err != nil {
		t.Fatalf("WriteFile(commit.sig): %v", err)
	}

	return parsed.AppendPayload(nil), allowedSignersPath, signaturePath
}

func sshVerify(
	t *testing.T,
	payload []byte,
	allowedSignersPath string,
	signaturePath string,
) ([]byte, error) {
	t.Helper()

	cmd := exec.CommandContext(
		t.Context(),
		"ssh-keygen",
		"-Y", "verify",
		"-n", "git",
		"-f", allowedSignersPath,
		"-I", signerPrincipal,
		"-s", signaturePath,
	) //#nosec G204
	cmd.Stdin = bytes.NewReader(payload)

	return cmd.CombinedOutput() //nolint:wrapcheck
}

func TestSSHSignedCommitVerifies(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			payload, allowedSignersPath, signaturePath := setupSSHSignedCommit(t, objectFormat)

			out, err := sshVerify(t, payload, allowedSignersPath, signaturePath)
			if err != nil {
				t.Fatalf("ssh-keygen verify failed: %v\n%s", err, out)
			}
		})
	}
}

func TestSSHSignedCommitRejectsTamperedPayload(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			payload, allowedSignersPath, signaturePath := setupSSHSignedCommit(t, objectFormat)
			payload = append([]byte(nil), payload...)
			payload[len(payload)-2] ^= 1

			out, err := sshVerify(t, payload, allowedSignersPath, signaturePath)
			if err == nil {
				t.Fatalf("ssh-keygen verify unexpectedly succeeded:\n%s", out)
			}
		})
	}
}
