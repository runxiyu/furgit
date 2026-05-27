package testgit

import (
	"strings"
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
)

// RevList runs "git rev-list" with args and parses one object ID per line.
func (testRepo *TestRepo) RevList(tb testing.TB, args ...string) []objectid.ObjectID {
	tb.Helper()

	cmdArgs := make([]string, 0, len(args)+1)
	cmdArgs = append(cmdArgs, "rev-list")
	cmdArgs = append(cmdArgs, args...)
	out := testRepo.Run(tb, cmdArgs...)

	lines := strings.Split(strings.TrimSpace(out), "\n")

	outIDs := make([]objectid.ObjectID, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		id, err := objectid.ParseHex(testRepo.algo, line)
		if err != nil {
			tb.Fatalf("parse rev-list oid %q: %v", line, err)
		}

		outIDs = append(outIDs, id)
	}

	return outIDs
}
