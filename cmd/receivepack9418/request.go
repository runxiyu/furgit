package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type gitProtoRequest struct {
	Command         string
	Pathname        string
	Host            string
	ExtraParameters []string
}

func parseGitProtoRequestPayload(payload []byte) (gitProtoRequest, error) {
	parts := bytes.Split(payload, []byte{0})
	if len(parts) == 0 || len(parts[0]) == 0 {
		return gitProtoRequest{}, errors.New("missing command/path segment")
	}

	commandPath := string(parts[0])
	command, pathname, ok := strings.Cut(commandPath, " ")
	if !ok || command == "" || pathname == "" {
		return gitProtoRequest{}, fmt.Errorf("malformed command/path segment %q", commandPath)
	}

	req := gitProtoRequest{
		Command:  command,
		Pathname: pathname,
	}

	i := 1
	if i < len(parts) && strings.HasPrefix(string(parts[i]), "host=") {
		req.Host = strings.TrimPrefix(string(parts[i]), "host=")
		i++
	}

	// No tail left.
	if i >= len(parts) {
		return req, nil
	}

	// If there is tail, grammar requires one empty field before extras.
	if len(parts[i]) != 0 {
		return gitProtoRequest{}, fmt.Errorf("unexpected token %q after host/path", string(parts[i]))
	}

	i++
	for ; i < len(parts); i++ {
		if len(parts[i]) == 0 {
			continue
		}

		req.ExtraParameters = append(req.ExtraParameters, string(parts[i]))
	}

	return req, nil
}
