package repository

import (
	"time"

	"codeberg.org/lindenii/furgit/config"
)

func detectPackedRefsTimeout(cfg *config.Config) time.Duration {
	timeoutValue, err := cfg.Lookup("core", "", "packedrefstimeout").Int()
	if err != nil {
		return time.Second
	}

	return time.Duration(timeoutValue) * time.Millisecond
}
