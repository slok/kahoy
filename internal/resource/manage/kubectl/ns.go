package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
)

// NamespaceEnsurerConfig is the configuration for NewNamespaceEnsurer.
type NamespaceEnsurerConfig struct {
	// Manager is the original manager used to apply and delete.
	Manager     manage.ResourceManager
	KubectlCmd  string
	KubeConfig  string
	KubeContext string
	CmdRunner   CmdRunner
	Out         io.Writer
	ErrOut      io.Writer
	Logger      log.Logger
}

func (c *NamespaceEnsurerConfig) defaults() error {
	if c.Manager == nil {
		return fmt.Errorf("resource manager is required")
	}

	if c.KubectlCmd == "" {
		c.KubectlCmd = "kubectl"
	}

	if c.Out == nil {
		c.Out = os.Stdout
	}

	if c.ErrOut == nil {
		c.ErrOut = os.Stderr
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "kubectl.NamespaceEnsurer"})

	if c.CmdRunner == nil {
		c.CmdRunner = newStdCmdRunner(c.Logger)
	}

	return nil
}

type namespaceEnsurer struct {
	manager     manage.ResourceManager
	kubectlCmd  string
	kubeConfig  string
	kubeContext string
	cmdRunner   CmdRunner
	out         io.Writer
	errOut      io.Writer
	logger      log.Logger
}

// NewNamespaceEnsurer returns a resource Manager based on Kubctl that will ensure
// the namespace of the applied resources are present before applying them.
func NewNamespaceEnsurer(config NamespaceEnsurerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return namespaceEnsurer{
		manager:     config.Manager,
		kubectlCmd:  config.KubectlCmd,
		kubeConfig:  config.KubeConfig,
		kubeContext: config.KubeContext,
		cmdRunner:   config.CmdRunner,
		out:         config.Out,
		errOut:      config.ErrOut,
		logger:      config.Logger,
	}, nil
}

func (n namespaceEnsurer) Apply(ctx context.Context, resources []model.Resource) error {
	namespaces := map[string]struct{}{}
	for _, r := range resources {
		ns := r.K8sObject.GetNamespace()
		if ns == "" {
			continue
		}

		namespaces[ns] = struct{}{}
	}

	// Ensure the Namespaces are present.
	for ns := range namespaces {
		err := n.ensureNamespace(ctx, ns)
		if err != nil {
			return fmt.Errorf("could not ensure namespace %q: %w", ns, err)
		}
	}

	return n.manager.Apply(ctx, resources)
}

func (n namespaceEnsurer) Delete(ctx context.Context, resources []model.Resource) error {
	return n.manager.Delete(ctx, resources)
}

func (n namespaceEnsurer) ensureNamespace(ctx context.Context, ns string) error {
	logger := n.logger.WithValues(log.Kv{"ext-cmd": "kubectl"})

	// Check if the ns is missing.
	{
		getArgs := newKubectlCmdArgs([]kubectlCmdOption{
			withGetCmd(),
			withContext(n.kubeContext),
			withConfig(n.kubeConfig),
			withNamespaceKind(),
			withResourceName(ns),
		})

		cmd := exec.CommandContext(ctx, n.kubectlCmd, getArgs...)
		var outErr bytes.Buffer
		cmd.Stderr = &outErr
		err := n.cmdRunner.Run(cmd)
		if err == nil {
			// Namespace already present.
			return nil
		}

		// Check if the error is of missing ns.
		// Errors of missing ns are in this way: `Error from server (NotFound): namespaces "xxxxxx" not found`.
		errorData := outErr.String()
		if !strings.Contains(strings.ToLower(errorData), "notfound") {
			return fmt.Errorf("could not get ns info: %s: %w", errorData, err)
		}
	}

	// Create the ns.
	{
		createArgs := newKubectlCmdArgs([]kubectlCmdOption{
			withCreateCmd(),
			withContext(n.kubeContext),
			withConfig(n.kubeConfig),
			withNamespaceKind(),
			withResourceName(ns),
		})

		var out, outErr bytes.Buffer
		cmd := exec.CommandContext(ctx, n.kubectlCmd, createArgs...)
		cmd.Stdout = &out
		cmd.Stderr = &outErr
		err := n.cmdRunner.Run(cmd)
		if err != nil {
			for _, line := range strings.Split(outErr.String(), "\n") {
				if line == "" {
					continue
				}
				logger.Errorf(line)
			}
			return fmt.Errorf("error while executing create namespace command: %s: %w", outErr.String(), err)
		}

		for _, line := range strings.Split(out.String(), "\n") {
			if line == "" {
				continue
			}
			logger.Infof(line)
		}
	}

	return nil
}
