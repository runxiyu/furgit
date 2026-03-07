package receivepack

import (
	protoreceive "codeberg.org/lindenii/furgit/protocol/v0v1/server/receivepack"
	"codeberg.org/lindenii/furgit/receivepack/service"
)

func translateResult(result *service.Result) protoreceive.ReportStatusResult {
	out := protoreceive.ReportStatusResult{
		UnpackError: result.UnpackError,
		Commands:    make([]protoreceive.CommandResult, 0, len(result.Commands)),
	}

	for _, command := range result.Commands {
		out.Commands = append(out.Commands, protoreceive.CommandResult{
			Name:         command.Name,
			Error:        command.Error,
			RefName:      command.RefName,
			OldID:        command.OldID,
			NewID:        command.NewID,
			ForcedUpdate: command.ForcedUpdate,
		})
	}

	return out
}
