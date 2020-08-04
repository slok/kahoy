package main

import (
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/util/homedir"
)

// Commandline subcommands IDs.
const (
	CmdArgApply = "apply"
)

// CmdConfig is the configuration of the command
type CmdConfig struct {
	// Command is the loaded command.
	Command string

	// Global is the configuration shared by all commands.
	Global struct {
		Debug bool
	}

	// Apply is the apply command configuration.
	Apply struct {
		KubeContext string
		KubeConfig  string
	}
}

// NewCmdConfig returns the application
func NewCmdConfig(args []string) (*CmdConfig, error) {
	kubeHome := filepath.Join(homedir.HomeDir(), ".kube", "config")

	c := CmdConfig{}
	app := kingpin.New("kahoy", "A simple Kubernetes deployer tool for raw manifests")
	app.Version(Version)
	app.DefaultEnvars()

	// Global.
	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Global.Debug)

	// Apply command.
	apply := app.Command(CmdArgApply, "Will take all the manifests in the directory and apply to a Kubernetes cluster.")
	apply.Flag("kube-config", "kubernetes configuration configuration path.").Default(kubeHome).StringVar(&c.Apply.KubeConfig)
	apply.Flag("kube-context", "kubernetes configuration context.").StringVar(&c.Apply.KubeContext)

	// Parse the commandline.
	cmd, err := app.Parse(args)
	if err != nil {
		return nil, err
	}
	c.Command = cmd

	return &c, nil
}
