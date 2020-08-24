package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/slok/kahoy/internal/log"
)

var (
	// Version is the app version.
	Version = "dev"
)

// GlobalConfig is the configuration shared by all the commands.
type GlobalConfig struct {
	Logger log.Logger
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run runs the main application.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// Cmd configuration.
	config, err := NewCmdConfig(args[1:])
	if err != nil {
		return fmt.Errorf("could not load command configuration: %w", err)
	}

	// Set up logger.
	var logger log.Logger = log.Noop
	if !config.Global.NoLog {
		// If not logger disabled use logrus logger.
		logrusLog := logrus.New()
		logrusLog.Out = stderr // By default logger goes to stderr (so it can split stdout prints).
		logrusLogEntry := logrus.NewEntry(logrusLog)
		logrusLogEntry.Logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   !config.Global.NoColor,
			DisableColors: config.Global.NoColor,
		})
		if config.Global.Debug {
			logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
		}
		logger = log.NewLogrus(logrusLogEntry)
	}

	logger = logger.WithValues(log.Kv{
		"app":     "kahoy",
		"version": Version,
	})
	logger.Debugf("debug level is enabled") // Will log only when debug enabled.

	// Global configuration.
	gConfig := GlobalConfig{
		Logger: logger,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	// Mapping for each command func and select the correct one.
	commands := map[string]func(ctx context.Context, config CmdConfig, globalConfig GlobalConfig) error{
		CmdArgApply: RunApply,
	}
	cmd, ok := commands[config.Command]
	if !ok {
		return fmt.Errorf("command %q is not valid", config.Command)
	}

	// Run the command.
	err = cmd(ctx, *config, gConfig)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	err := Run(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
