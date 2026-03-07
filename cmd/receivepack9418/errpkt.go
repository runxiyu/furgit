package main

import (
	"io"

	"codeberg.org/lindenii/furgit/format/pktline"
)

func writeErrPkt(w io.Writer, message string) {
	payload := []byte("ERR " + message + "\n")

	frame, err := pktline.AppendData(nil, payload)
	if err != nil {
		return
	}

	_, _ = w.Write(frame)
}
