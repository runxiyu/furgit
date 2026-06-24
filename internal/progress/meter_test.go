package progress_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"lindenii.org/go/furgit/internal/progress"
	"lindenii.org/go/lgo/iowrap"
)

func TestMeterConcurrentAdd(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	meter := progress.New(progress.Options{
		Writer: iowrap.NopFlush(&buf),
		Title:  "test",
		Total:  1000,
	})

	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 100 {
				meter.Add(1, 0)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	meter.Stop("done")

	if got := buf.String(); !strings.Contains(got, "100% (1000/1000)") {
		t.Fatalf("final line = %q, want it to contain %q", got, "100% (1000/1000)")
	}
}
