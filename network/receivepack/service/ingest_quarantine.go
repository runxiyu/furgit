package service

import (
	"lindenii.org/go/furgit/internal/utils"
	objectstore "lindenii.org/go/furgit/object/store"
)

func (service *Service) ingestQuarantine(
	result *Result,
	commands []Command,
	req *Request,
) (objectstore.Quarantine, bool) {
	if !req.PackExpected {
		return nil, true
	}

	if req.Pack == nil {
		utils.BestEffortFprintf(service.opts.Progress, "unpack failed: missing pack stream.\n")

		result.UnpackError = "missing pack stream"
		fillCommandErrors(result, commands, "missing pack stream")

		return nil, false
	}

	if service.opts.ObjectIngress == nil {
		utils.BestEffortFprintf(service.opts.Progress, "unpack failed: object ingress not configured.\n")

		result.UnpackError = "object ingress not configured"
		fillCommandErrors(result, commands, "object ingress not configured")

		return nil, false
	}

	var err error

	err = service.opts.ExistingObjects.Refresh()
	if err != nil {
		utils.BestEffortFprintf(service.opts.Progress, "unpack failed: refresh existing objects: %v.\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		return nil, false
	}

	utils.BestEffortFprintf(service.opts.Progress, "creating quarantine...\r")

	quarantine, err := service.opts.ObjectIngress.BeginQuarantine(objectstore.QuarantineOptions{})
	if err != nil {
		utils.BestEffortFprintf(service.opts.Progress, "unpack failed: %v.\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		return nil, false
	}

	utils.BestEffortFprintf(service.opts.Progress, "creating quarantine: done.\n")
	utils.BestEffortFprintf(service.opts.Progress, "unpacking...\r")

	err = quarantine.WritePack(req.Pack, objectstore.PackWriteOptions{
		ThinBase:           service.opts.ExistingObjects,
		Progress:           service.opts.Progress,
		RequireTrailingEOF: false,
	})
	if err != nil {
		utils.BestEffortFprintf(service.opts.Progress, "unpack failed: %v.\n", err)

		result.UnpackError = err.Error()
		fillCommandErrors(result, commands, err.Error())

		_ = quarantine.Discard()

		return nil, false
	}

	utils.BestEffortFprintf(service.opts.Progress, "unpacking: done.\n")

	return quarantine, true
}
