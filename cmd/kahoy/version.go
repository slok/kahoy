package main

import (
	"context"
	"fmt"
)

// RunVersion runs version command.
func RunVersion(_ context.Context, _ CmdConfig, globalConfig GlobalConfig) error {
	fmt.Fprintln(globalConfig.Stdout, Version)

	return nil
}
