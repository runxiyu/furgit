package receivepack

import (
	"crypto/rand"
)

func defaultAgent() string {
	return "furgit"
}

func defaultSessionID() string {
	return "furgit-" + rand.Text()
}

func defaultPushCertNonce() string {
	return "furgit-" + rand.Text()
}
