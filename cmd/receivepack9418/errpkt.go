package main

import (
	"io"

	"lindenii.org/go/furgit/network/protocol/pktline"
)

func writeErrPkt(w io.Writer, message string) {
	payload := []byte("ERR " + message + "\n")

	frame, err := pktline.AppendData(nil, payload)
	if err != nil {
		return
	}

	_, _ = w.Write(frame)
}
