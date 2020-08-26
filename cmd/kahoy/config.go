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

// Defaults.
const (
	DefaultConfigFile = "kahoy.yml"
)

// Apply modes.
const (
	ApplyModePaths = "paths"
	ApplyModeGit   = "git"
)

// CmdConfig is the configuration of the command
type CmdConfig struct {
	// Command is the loaded command.
	Command string

	// Global is the configuration shared by all commands.
	Global struct {
		Debug      bool
		NoColor    bool
		NoLog      bool
		ConfigFile string
	}

	// Apply is the apply command configuration.
	Apply struct {
		KubeContext              string
		KubeConfig               string
		ManifestsPathOld         string
		ManifestsPathNew         string
		DiffMode                 bool
		ExcludeManifests         []string
		IncludeManifests         []string
		ExcludeKubeTypeResources []string
		GitBeforeCommit          string
		GitDefaultBranch         string
		GitDiffFilter            bool
		Mode                     string
		DryRun                   bool
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
	app.Flag("config-file", "App configuration file.").Default(DefaultConfigFile).StringVar(&c.Global.ConfigFile)

	// Apply command.
	apply := app.Command(CmdArgApply, "Will take all the manifests in the directory and apply to a Kubernetes cluster.")
	apply.Flag("kube-config", "kubernetes configuration configuration path.").Default(kubeHome).StringVar(&c.Apply.KubeConfig)
	apply.Flag("kube-context", "kubernetes configuration context.").StringVar(&c.Apply.KubeContext)
	apply.Flag("diff", "diff instead of applying changes.").BoolVar(&c.Apply.DiffMode)
	apply.Flag("dry-run", "execute in dry-run, is safe, can be run without Kubernetes cluster.").BoolVar(&c.Apply.DryRun)
	apply.Flag("mode", "selects how apply will select the state, load manifests... git needs to be executed from a git repository.").Default(ApplyModeGit).EnumVar(&c.Apply.Mode, ApplyModePaths, ApplyModeGit)
	apply.Flag("fs-old-manifests-path", "kubernetes current manifests path.").Required().StringVar(&c.Apply.ManifestsPathOld)
	apply.Flag("fs-new-manifests-path", "kubernetes expected manifests path.").Required().StringVar(&c.Apply.ManifestsPathNew)
	apply.Flag("fs-exclude", "regex to ignore manifest files and dirs. Can be repeated.").StringsVar(&c.Apply.ExcludeManifests)
	apply.Flag("fs-include", "regex to include manifest files and dirs, everything else will be ignored. Exclude has preference. Can be repeated.").StringsVar(&c.Apply.IncludeManifests)
	apply.Flag("git-diff-filter", "excludes everything except the files changed in before-commit and HEAD git diff.").BoolVar(&c.Apply.GitDiffFilter)
	apply.Flag("git-before-commit-sha", "the git hash used as the old state to get the apply/delete plan, if not passed, it will search using merge-base common ancestor of current HEAD and default branch.").StringVar(&c.Apply.GitBeforeCommit)
	apply.Flag("git-default-branch", "git repository default branch. Used to search common parent (default-branch and HEAD) when 'before-commit' not provided. Only supports local branches (no remote branches, tags, hashes...).").Default("master").StringVar(&c.Apply.GitDefaultBranch)
	apply.Flag("kube-exclude-type", "regex to ignore Kubernetes resources by api version and type (apps/v1/Deployment, v1/Pod...). Can be repeated.").StringsVar(&c.Apply.ExcludeKubeTypeResources)

	// Parse the commandline.
	cmd, err := app.Parse(args)
	if err != nil {
		return nil, err
	}
	c.Command = cmd

	return &c, nil
}
