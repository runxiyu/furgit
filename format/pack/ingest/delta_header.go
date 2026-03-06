package ingest

import deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"

// finalizeStreamPackHash consumes trailer bytes and verifies stream integrity.
// readDeltaHeaderSizes reads source and destination sizes from one delta payload.
func readDeltaHeaderSizes(payload []byte) (int, int, error) {
	reader := &byteSliceReader{data: payload}

	return deltaapply.ReadHeaderSizes(reader)
}
