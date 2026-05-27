package receivepack

import (
	"strings"

	common "lindenii.org/go/furgit/network/protocol/v0v1/server"
)

func parseVersion(gitProtocol string) common.Version {
	if gitProtocol == "" {
		return common.Version0
	}

	var highestRequested uint8

	for field := range strings.SplitSeq(gitProtocol, ":") {
		switch field {
		case "version=0":
		case "version=1":
			if highestRequested < 1 {
				highestRequested = 1
			}
		case "version=2":
			if highestRequested < 2 {
				highestRequested = 2
			}
		}
	}

	if highestRequested == 1 {
		return common.Version1
	}

	return common.Version0
}
