package testgit

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// packObjectsReadCloser wraps a pipe reader and process wait fn.
type packObjectsReadCloser struct {
	reader io.ReadCloser
	wait   func() error
	once   sync.Once
}

// Read proxies reads to the wrapped reader.
func (reader *packObjectsReadCloser) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst)
}

// Close closes the stream and waits for the underlying process.
func (reader *packObjectsReadCloser) Close() error {
	var out error

	reader.once.Do(func() {
		errClose := reader.reader.Close()
		errWait := reader.wait()

		if errClose != nil {
			out = errClose

			return
		}

		out = errWait
	})

	return out
}

// PackObjectsReader streams `git pack-objects --stdout --revs` output.
func (testRepo *TestRepo) PackObjectsReader(tb testing.TB, revs []string, thin bool) io.ReadCloser {
	tb.Helper()

	args := []string{"pack-objects", "--stdout", "--revs"}
	if thin {
		args = append(args, "--thin")
	}

	//nolint:noctx
	cmd := exec.Command("git", args...) //#nosec G204
	cmd.Dir = testRepo.dir
	cmd.Env = testRepo.env
	cmd.Stdin = strings.NewReader(strings.Join(revs, "\n") + "\n")

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	stderr := &strings.Builder{}
	cmd.Stderr = stderr

	waitDone := make(chan error, 1)

	go func() {
		err := cmd.Start()
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("git %v start failed: %w", args, err))

			waitDone <- nil

			return
		}

		err = cmd.Wait()
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("git %v failed: %w\n%s", args, err, stderr.String()))
		} else {
			_ = pw.Close()
		}

		waitDone <- nil
	}()

	return &packObjectsReadCloser{
		reader: pr,
		wait: func() error {
			<-waitDone

			return nil
		},
	}
}
