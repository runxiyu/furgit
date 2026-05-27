package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"lindenii.org/go/furgit/network/receivepack"
	objectdual "lindenii.org/go/furgit/object/store/dual"
	objectloose "lindenii.org/go/furgit/object/store/loose"
	objectpacked "lindenii.org/go/furgit/object/store/packed"
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

	objectIngress, cleanupObjectIngress, err := srv.openObjectIngress()
	if err != nil {
		writeErrPkt(writer, fmt.Sprintf("object ingress unavailable: %v", err))
		_ = writer.Flush()

		log.Printf("receivepack9418: %s: object ingress unavailable: %v", conn.RemoteAddr(), err)

		return
	}

	defer cleanupObjectIngress()

	opts := receivepack.Options{
		GitProtocol:     gitProtocol,
		Algorithm:       srv.repo.Algorithm(),
		Refs:            srv.repo.RefStore(),
		ExistingObjects: srv.repo.ObjectStore(),
		ObjectIngress:   objectIngress,
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

func (srv *server) openObjectIngress() (*objectdual.Dual, func(), error) {
	err := srv.objectsRoot.Mkdir("pack", 0o755)
	if err != nil && !os.IsExist(err) {
		return nil, nil, err
	}

	packRoot, err := srv.objectsRoot.OpenRoot("pack")
	if err != nil {
		return nil, nil, err
	}

	looseStore, err := objectloose.New(srv.objectsRoot, srv.repo.Algorithm())
	if err != nil {
		_ = packRoot.Close()

		return nil, nil, err
	}

	packedStore, err := objectpacked.New(packRoot, srv.repo.Algorithm(), objectpacked.Options{WriteRev: true})
	if err != nil {
		_ = looseStore.Close()
		_ = packRoot.Close()

		return nil, nil, err
	}

	cleanup := func() {
		_ = packedStore.Close()
		_ = looseStore.Close()
		_ = packRoot.Close()
	}

	return objectdual.New(looseStore, packedStore), cleanup, nil
}
