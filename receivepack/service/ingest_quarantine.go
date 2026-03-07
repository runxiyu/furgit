package service

import (
	"os"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
)

func (service *Service) ingestQuarantine(
	result *Result,
	commands []Command,
	req *Request,
) (string, *os.Root, bool) {
	if !req.PackExpected {
		return "", nil, true
	}

	if req.Pack == nil {
		result.UnpackError = "missing pack stream"
		fillCommandErrors(result, commands, "missing pack stream")

		return "", nil, false
	}

	if service.opts.ObjectsRoot == nil {
		result.UnpackError = "objects root not configured"
		fillCommandErrors(result, commands, "objects root not configured")

		return "", nil, false
	}

	quarantineName, quarantineRoot, err := service.createQuarantineRoot()
	if err != nil {
		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		return "", nil, false
	}

	quarantinePackRoot, err := service.openQuarantinePackRoot(quarantineRoot)
	if err != nil {
		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		_ = quarantineRoot.Close()
		_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)

		return "", nil, false
	}

	ingested, err := ingest.Ingest(
		req.Pack,
		quarantinePackRoot,
		service.opts.Algorithm,
		true,
		true,
		service.opts.ExistingObjects,
	)

	_ = quarantinePackRoot.Close()

	if err != nil {
		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		_ = quarantineRoot.Close()
		_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)

		return "", nil, false
	}

	result.Ingest = &ingested

	return quarantineName, quarantineRoot, true
}
