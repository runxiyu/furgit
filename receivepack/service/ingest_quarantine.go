package service

import (
	"os"

	"codeberg.org/lindenii/furgit/format/pack/ingest"
	"codeberg.org/lindenii/furgit/internal/utils"
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
		utils.WriteProgressf(service.opts.Progress, "unpack failed: missing pack stream\n")

		result.UnpackError = "missing pack stream"
		fillCommandErrors(result, commands, "missing pack stream")

		return "", nil, false
	}

	if service.opts.ObjectsRoot == nil {
		utils.WriteProgressf(service.opts.Progress, "unpack failed: objects root not configured\n")

		result.UnpackError = "objects root not configured"
		fillCommandErrors(result, commands, "objects root not configured")

		return "", nil, false
	}

	utils.WriteProgressf(service.opts.Progress, "creating quarantine...\r")
	quarantineName, quarantineRoot, err := service.createQuarantineRoot()
	if err != nil {
		utils.WriteProgressf(service.opts.Progress, "unpack failed: %v\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		return "", nil, false
	}

	quarantinePackRoot, err := service.openQuarantinePackRoot(quarantineRoot)
	if err != nil {
		utils.WriteProgressf(service.opts.Progress, "unpack failed: %v\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		_ = quarantineRoot.Close()
		_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)

		return "", nil, false
	}
	utils.WriteProgressf(service.opts.Progress, "creating quarantine: done.\n")
	utils.WriteProgressf(service.opts.Progress, "unpacking...\r")

	ingested, err := ingest.Ingest(
		req.Pack,
		quarantinePackRoot,
		service.opts.Algorithm,
		ingest.Options{
			FixThin:  true,
			WriteRev: true,
			Base:     service.opts.ExistingObjects,
			Progress: service.opts.Progress,
		},
	)

	_ = quarantinePackRoot.Close()

	if err != nil {
		utils.WriteProgressf(service.opts.Progress, "unpack failed: %v\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		_ = quarantineRoot.Close()
		_ = service.opts.ObjectsRoot.RemoveAll(quarantineName)

		return "", nil, false
	}

	utils.WriteProgressf(service.opts.Progress, "unpacking: done (%d objects, %s).\n", ingested.ObjectCount, ingested.PackHash)

	result.Ingest = &ingested

	return quarantineName, quarantineRoot, true
}
