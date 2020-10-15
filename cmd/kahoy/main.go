package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/sirupsen/logrus"

	"github.com/slok/kahoy/internal/configuration"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

var (
	// Version is the app version.
	Version = "dev"
)

// GlobalConfig is the configuration shared by all the commands.
type GlobalConfig struct {
	AppConfig model.AppConfig
	Logger    log.Logger
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
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

	// Load app configuration if present.
	var appConfig model.AppConfig
	configData, err := ioutil.ReadFile(config.Global.ConfigFile)
	if err != nil {
		// If the default configuration file is missing, don't fail, if is a custom one fail.
		if config.Global.ConfigFile != DefaultConfigFile || !os.IsNotExist(err) {
			return fmt.Errorf("could not load %q config file: %w", config.Global.ConfigFile, err)
		}
	} else {
		cfg, err := configuration.NewYAMLV1Loader(string(configData)).Load(ctx)
		if err != nil {
			return fmt.Errorf("could not load app configuration: %w", err)
		}
		appConfig = *cfg
		logger.WithValues(log.Kv{"config-file": config.Global.ConfigFile}).Infof("app configuration loaded")
	}

	err = appConfig.Validate(ctx)
	if err != nil {
		return fmt.Errorf("invalid app configuration: %w", err)
	}

	// Set global configuration.
	gConfig := GlobalConfig{
		AppConfig: appConfig,
		Logger:    logger,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
	}

	// Prepare our run entrypoints.
	var g run.Group

	// Cmd run.
	{
		// Mapping for each command func and select the correct one.
		commands := map[string]func(ctx context.Context, config CmdConfig, globalConfig GlobalConfig) error{
			CmdArgApply:   RunApply,
			CmdArgVersion: RunVersion,
		}
		cmd, ok := commands[config.Command]
		if !ok {
			return fmt.Errorf("command %q is not valid", config.Command)
		}

		ctx, cancel := context.WithCancel(ctx)
		g.Add(
			func() error {
				return cmd(ctx, *config, gConfig)
			},
			func(_ error) {
				logger.Debugf("stopping cmd execution")
				cancel()
			},
		)
	}

	// OS signals.
	{
		sigC := make(chan os.Signal, 1)
		exitC := make(chan struct{})
		signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)

		g.Add(
			func() error {
				select {
				case s := <-sigC:
					logger.Infof("signal %q received", s)
					return nil
				case <-exitC:
					return nil
				}
			},
			func(_ error) {
				close(exitC)
			},
		)
	}

	return g.Run()
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
