package files

import (
	"errors"
	"os"
	"time"
)

func (tx *Transaction) createPackedLock(timeout time.Duration) error {
	const (
		initialBackoffMs     = 1
		backoffMaxMultiplier = 1000
	)

	deadline := time.Now().Add(timeout)
	multiplier := 1
	n := 1

	for {
		file, err := tx.store.commonRoot.OpenFile("packed-refs.lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return file.Close()
		}

		if !errors.Is(err, os.ErrExist) {
			return err
		}

		if timeout == 0 || (timeout > 0 && time.Now().After(deadline)) {
			return err
		}

		backoffMs := multiplier * initialBackoffMs
		waitMs := (750 + tx.store.lockRand.Intn(500)) * backoffMs / 1000
		time.Sleep(time.Duration(waitMs) * time.Millisecond)

		multiplier += 2*n + 1
		if multiplier > backoffMaxMultiplier {
			multiplier = backoffMaxMultiplier
		} else {
			n++
		}
	}
}
