package main

import (
	"context"

	"github.com/slok/kahoy/internal/log"
)

// RunApply runs the apply command.
func RunApply(ctx context.Context, cmdConfig CmdConfig, globalConfig GlobalConfig) error {
	logger := globalConfig.Logger.WithValues(log.Kv{"cmd": "apply"})
	logger.Infof("running command")

	return nil
}
