package receivepack

import (
	protoreceive "lindenii.org/go/furgit/network/protocol/v0v1/server/receivepack"
	"lindenii.org/go/furgit/network/receivepack/service"
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
