package receivepack

import (
	protoreceive "codeberg.org/lindenii/furgit/protocol/v0v1/server/receivepack"
	"codeberg.org/lindenii/furgit/receivepack/internal/service"
)

func translateCommands(commands []protoreceive.Command) []service.Command {
	out := make([]service.Command, 0, len(commands))
	for _, command := range commands {
		out = append(out, service.Command{
			OldID: command.OldID,
			NewID: command.NewID,
			Name:  command.Name,
		})
	}

	return out
}

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
