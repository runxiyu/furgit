package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"codeberg.org/lindenii/furgit/network/receivepack"
)

func (srv *server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	req, err := readGitProtoRequest(reader)
	if err != nil {
		writeErrPkt(writer, fmt.Sprintf("invalid initial request: %v", err))
		_ = writer.Flush()

		log.Printf("receivepack9418: %s: invalid initial request: %v", conn.RemoteAddr(), err)

		return
	}

	if req.Command != "git-receive-pack" {
		writeErrPkt(writer, fmt.Sprintf("unsupported command %q", req.Command))
		_ = writer.Flush()

		log.Printf("receivepack9418: %s: unsupported command %q", conn.RemoteAddr(), req.Command)

		return
	}

	gitProtocol := strings.Join(req.ExtraParameters, ":")

	opts := receivepack.Options{
		GitProtocol:     gitProtocol,
		Algorithm:       srv.repo.Algorithm(),
		Refs:            srv.repo.Refs(),
		ExistingObjects: srv.repo.Objects(),
		ObjectsRoot:     srv.objectsRoot,
	}

	err = receivepack.ReceivePack(context.Background(), writer, reader, opts)
	if err != nil {
		_ = writer.Flush()

		log.Printf(
			"receivepack9418: %s: receive-pack failed (path=%q host=%q extras=%v): %v",
			conn.RemoteAddr(),
			req.Pathname,
			req.Host,
			req.ExtraParameters,
			err,
		)

		return
	}

	err = writer.Flush()
	if err != nil {
		log.Printf("receivepack9418: %s: flush failed: %v", conn.RemoteAddr(), err)

		return
	}
}
