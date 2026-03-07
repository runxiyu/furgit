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
