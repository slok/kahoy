package main

import (
	"fmt"
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
		Mode                     string
		DryRun                   bool
		IncludeChanges           bool
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
	apply.Flag("kube-config", "Kubernetes configuration configuration path.").Envar("KUBECONFIG").Default(kubeHome).StringVar(&c.Apply.KubeConfig)
	apply.Flag("kube-context", "Kubernetes configuration context.").StringVar(&c.Apply.KubeContext)
	apply.Flag("diff", "Diff instead of applying changes.").BoolVar(&c.Apply.DiffMode)
	apply.Flag("dry-run", "Execute in dry-run, is safe, can be run without Kubernetes cluster.").BoolVar(&c.Apply.DryRun)
	apply.Flag("mode", "Selects how apply will select the state, load manifests... git needs to be executed from a git repository.").Default(ApplyModeGit).EnumVar(&c.Apply.Mode, ApplyModePaths, ApplyModeGit)
	apply.Flag("fs-old-manifests-path", "Kubernetes current manifests path.").Short('o').StringVar(&c.Apply.ManifestsPathOld)
	apply.Flag("fs-new-manifests-path", "Kubernetes expected manifests path.").Short('n').Required().StringVar(&c.Apply.ManifestsPathNew)
	apply.Flag("fs-exclude", "Regex to ignore manifest files and dirs. Can be repeated.").Short('e').StringsVar(&c.Apply.ExcludeManifests)
	apply.Flag("fs-include", "Regex to include manifest files and dirs, everything else will be ignored. Exclude has preference. Can be repeated.").Short('i').StringsVar(&c.Apply.IncludeManifests)
	apply.Flag("git-before-commit-sha", "The git hash used as the old state to get the apply/delete plan, if not passed, it will search using merge-base common ancestor of current HEAD and default branch.").Short('c').StringVar(&c.Apply.GitBeforeCommit)
	apply.Flag("git-default-branch", "Git repository default branch. Used to search common parent (default-branch and HEAD) when 'before-commit' not provided. Only supports local branches (no remote branches, tags, hashes...).").Default("master").StringVar(&c.Apply.GitDefaultBranch)
	apply.Flag("kube-exclude-type", "Regex to ignore Kubernetes resources by api version and type (apps/v1/Deployment, v1/Pod...). Can be repeated.").Short('a').StringsVar(&c.Apply.ExcludeKubeTypeResources)
	apply.Flag("include-changes", "Excludes all the resources without changes (old vs new states).").Short('f').BoolVar(&c.Apply.IncludeChanges)

	// Deprecated flags.
	apply.Flag("git-diff-filter", "DEPRECATED, use --include-changes.").Hidden().BoolVar(&c.Apply.IncludeChanges)

	// Parse the commandline.
	cmd, err := app.Parse(args)
	if err != nil {
		return nil, err
	}
	c.Command = cmd

	err = c.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid cmd configuration: %w", err)
	}

	return &c, nil
}

func (c *CmdConfig) validate() error {
	if c.Apply.DryRun && c.Apply.DiffMode {
		return fmt.Errorf(`only one of "dry run" and "diff" execution modes can be used at the same time`)
	}

	switch c.Apply.Mode {
	case ApplyModePaths:
		if c.Apply.ManifestsPathOld == "" {
			return fmt.Errorf("manifests old path is required when using %q mode", ApplyModePaths)
		}

	case ApplyModeGit:
		if c.Apply.ManifestsPathOld == "" {
			c.Apply.ManifestsPathOld = c.Apply.ManifestsPathNew
		}

		if c.Apply.GitDefaultBranch == "" && c.Apply.GitBeforeCommit == "" {
			return fmt.Errorf(`at least one of "git default branch" or "git before commit" is required`)
		}
	default:
		return fmt.Errorf("unknown mode: %q", c.Apply.Mode)
	}

	return nil
}
