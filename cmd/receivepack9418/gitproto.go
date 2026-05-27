package main

import (
	"fmt"
	"io"

	"lindenii.org/go/furgit/network/protocol/pktline"
)

func readGitProtoRequest(r io.Reader) (gitProtoRequest, error) {
	dec := pktline.NewDecoder(r, pktline.ReadOptions{})

	frame, err := dec.ReadFrame()
	if err != nil {
		return gitProtoRequest{}, err
	}

	if frame.Type != pktline.PacketData {
		return gitProtoRequest{}, fmt.Errorf("expected initial pkt-line data, got %v", frame.Type)
	}

	return parseGitProtoRequestPayload(frame.Payload)
}
