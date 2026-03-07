package receivepack

import (
	"context"
	"io"

	"codeberg.org/lindenii/furgit/format/pktline"
	common "codeberg.org/lindenii/furgit/protocol/v0v1/server"
	protoreceive "codeberg.org/lindenii/furgit/protocol/v0v1/server/receivepack"
	"codeberg.org/lindenii/furgit/receivepack/internal/service"
)

// ReceivePack serves one receive-pack session over r/w.
func ReceivePack(
	ctx context.Context,
	w pktline.WriteFlusher,
	r io.Reader,
	opts Options,
) error {
	err := validateOptions(opts)
	if err != nil {
		return err
	}

	version := parseVersion(opts.GitProtocol)

	base := common.NewSession(r, w, common.Options{
		Version:   version,
		Algorithm: opts.Algorithm,
	})

	protoSession := protoreceive.NewSession(base, protoreceive.Capabilities{
		ReportStatus:   true,
		ReportStatusV2: true,
		DeleteRefs:     true,
		SideBand64K:    true,
		Quiet:          true,
		Atomic:         true,
		OfsDelta:       true,
		PushOptions:    true,
		ObjectFormat:   opts.Algorithm,
		// TODO: PushCertNonce, SessionID, Agent, whatever.
	})

	refs, err := advertisedRefs(opts)
	if err != nil {
		return err
	}

	err = protoSession.AdvertiseRefs(common.Advertisement{Refs: refs})
	if err != nil {
		return err
	}

	req, err := protoSession.ReadRequest()
	if err != nil {
		return err
	}

	serviceReq := &service.Request{
		Commands:     translateCommands(req.Commands),
		PushOptions:  append([]string(nil), req.PushOptions...),
		Atomic:       req.Capabilities.Atomic,
		DeleteOnly:   req.DeleteOnly,
		PackExpected: req.PackExpected,
		Pack:         r,
	}

	svc := service.New(service.Options{
		Algorithm:       opts.Algorithm,
		Refs:            opts.Refs,
		ExistingObjects: opts.ExistingObjects,
		ObjectsRoot:     opts.ObjectsRoot,
		PromotedObjectPermissions: translatePromotedObjectPermissions(
			opts.PromotedObjectPermissions,
		),
	})

	result, err := svc.Execute(ctx, serviceReq)
	if err != nil {
		return err
	}

	protoResult := translateResult(result)

	if req.Capabilities.ReportStatusV2 {
		return protoSession.WriteReportStatusV2(protoResult)
	}

	if req.Capabilities.ReportStatus {
		return protoSession.WriteReportStatus(protoResult)
	}

	return nil
}

func translatePromotedObjectPermissions(
	perms *PromotedObjectPermissions,
) *service.PromotedObjectPermissions {
	if perms == nil {
		return nil
	}

	return &service.PromotedObjectPermissions{
		DirMode:  perms.DirMode,
		FileMode: perms.FileMode,
	}
}
