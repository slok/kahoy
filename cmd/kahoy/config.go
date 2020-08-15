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
		Debug   bool
		NoColor bool
		NoLog   bool
	}

	// Apply is the apply command configuration.
	Apply struct {
		KubeContext      string
		KubeConfig       string
		ManifestsPathOld string
		ManifestsPathNew string
		DiffMode         bool
		IgnoreManifests  []string
		DryRun           bool
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
	app.Flag("no-color", "Disable color.").BoolVar(&c.Global.NoColor)
	app.Flag("no-log", "Disable logger.").BoolVar(&c.Global.NoLog)

	// Apply command.
	apply := app.Command(CmdArgApply, "Will take all the manifests in the directory and apply to a Kubernetes cluster.")
	apply.Flag("kube-config", "kubernetes configuration configuration path.").Default(kubeHome).StringVar(&c.Apply.KubeConfig)
	apply.Flag("kube-context", "kubernetes configuration context.").StringVar(&c.Apply.KubeContext)
	apply.Flag("diff", "diff instead of applying changes.").BoolVar(&c.Apply.DiffMode)
	apply.Flag("fs-old-manifests-path", "kubernetes current manifests path.").Required().StringVar(&c.Apply.ManifestsPathOld)
	apply.Flag("fs-new-manifests-path", "kubernetes expected manifests path.").Required().StringVar(&c.Apply.ManifestsPathNew)
	apply.Flag("fs-ignore", "regex to ignore manifest files and dirs (can be repeated).").StringsVar(&c.Apply.IgnoreManifests)
	apply.Flag("dry-run", "execute in dry-run, is safe, can be run without Kubernetes cluster.").BoolVar(&c.Apply.DryRun)

	// Parse the commandline.
	cmd, err := app.Parse(args)
	if err != nil {
		return nil, err
	}
	c.Command = cmd

	return &c, nil
}
