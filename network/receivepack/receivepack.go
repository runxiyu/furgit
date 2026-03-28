package receivepack

import (
	"context"
	"io"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
	common "codeberg.org/lindenii/furgit/network/protocol/v0v1/server"
	protoreceive "codeberg.org/lindenii/furgit/network/protocol/v0v1/server/receivepack"
	"codeberg.org/lindenii/furgit/network/receivepack/service"
)

// TODO: Some more designing to do. In particular, we'd like to have access to
// commit graphs and stored object abstractions and such here, especially because
// hooks might want to access full repos, but we risk creating
// circular dependencies if we import repository/ here. Might need an interface-ish
// design, but that risks being over-complicated.
// Theoretically we could also just give the hooks an os.Root but that
// feels a bit ugly.

// ReceivePack serves one receive-pack session over r/w.
//
// Labels: Deps-Borrowed.
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

	agent := opts.Agent
	if agent == "" {
		agent = defaultAgent()
	}

	sessionID := opts.SessionID
	if sessionID == "" {
		sessionID = defaultSessionID()
	}

	pushCertNonce := opts.PushCertNonce
	if pushCertNonce == "" {
		pushCertNonce = defaultPushCertNonce()
	}

	protoSession := protoreceive.NewSession(base, protoreceive.Capabilities{
		ReportStatus:   true,
		ReportStatusV2: true,
		DeleteRefs:     true,
		SideBand64K:    true,
		Quiet:          true,
		Atomic:         true,
		OfsDelta:       true,
		PushOptions:    true,
		PushCertNonce:  pushCertNonce,
		SessionID:      sessionID,
		ObjectFormat:   opts.Algorithm,
		Agent:          agent,
	})

	refs, err := advertisedRefs(opts)
	if err != nil {
		return err
	}

	err = protoSession.AdvertiseRefs(common.Advertisement{Refs: refs})
	if err != nil {
		return err
	}

	err = base.FlushIO()
	if err != nil {
		return err
	}

	req, err := protoSession.ReadRequest()
	if err != nil {
		return err
	}

	progressWriter := protoSession.ProgressWriter()
	progressFlush := base.FlushIO

	if req.Capabilities.Quiet {
		progressWriter = io.Discard
		progressFlush = nil
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
		CommitGraph:     opts.CommitGraph,
		ObjectsRoot:     opts.ObjectsRoot,
		Progress:        progressWriter,
		ProgressFlush:   progressFlush,
		PromotedObjectPermissions: translatePromotedObjectPermissions(
			opts.PromotedObjectPermissions,
		),
		Hook: translateHook(opts.Hook),
		HookIO: service.HookIO{
			Progress: progressWriter,
			Error:    protoSession.ErrorWriter(),
		},
	})

	result, err := svc.Execute(ctx, serviceReq)
	if err != nil {
		return err
	}

	protoResult := translateResult(result)

	if req.Capabilities.ReportStatusV2 {
		err = protoSession.WriteReportStatusV2(protoResult)
		if err != nil {
			return err
		}
	} else if req.Capabilities.ReportStatus {
		err = protoSession.WriteReportStatus(protoResult)
		if err != nil {
			return err
		}
	}

	return base.FlushIO()
}
