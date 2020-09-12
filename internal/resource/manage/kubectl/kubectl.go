package kubectl

import (
	"bufio"
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

// ManagerConfig is the configuration for NewManager.
type ManagerConfig struct {
	KubectlCmd                string
	KubeConfig                string
	KubeContext               string
	KubeFieldManager          string
	DisableKubeForceConflicts bool
	YAMLEncoder               K8sObjectEncoder
	CmdRunner                 CmdRunner
	Out                       io.Writer
	ErrOut                    io.Writer
	Logger                    log.Logger
}

func (c *ManagerConfig) defaults() error {
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
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "kubectl.Manager"})

	if c.CmdRunner == nil {
		c.CmdRunner = newStdCmdRunner(c.Logger)
	}

	if c.YAMLEncoder == nil {
		return fmt.Errorf("yaml encoder is required")
	}

	return nil
}

type manager struct {
	kubectlCmd  string
	yamlEncoder K8sObjectEncoder
	cmdRunner   CmdRunner
	out         io.Writer
	errOut      io.Writer
	logger      log.Logger

	applyArgs  []string
	deleteArgs []string
}

// NewManager returns a resource Manager based on Kubctl that will apply changes.
func NewManager(config ManagerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	applyArgs := newKubectlCmdArgs([]kubectlCmdOption{
		withApplyCmd(),
		withContext(config.KubeContext),
		withConfig(config.KubeConfig),
		withForceConflicts(!config.DisableKubeForceConflicts),
		withFieldManager(config.KubeFieldManager),
		withServerSide(true),
		withStdIn(),
	})

	deleteArgs := newKubectlCmdArgs([]kubectlCmdOption{
		withDeleteCmd(),
		withContext(config.KubeContext),
		withConfig(config.KubeConfig),
		withIgnoreNotFound(true),
		withWait(false),
		withStdIn(),
	})

	return manager{
		kubectlCmd:  config.KubectlCmd,
		yamlEncoder: config.YAMLEncoder,
		cmdRunner:   config.CmdRunner,
		out:         config.Out,
		errOut:      config.ErrOut,
		logger:      config.Logger,
		applyArgs:   applyArgs,
		deleteArgs:  deleteArgs,
	}, nil
}

func (m manager) Apply(ctx context.Context, resources []model.Resource) error {
	err := m.execute(ctx, resources, m.applyArgs)
	if err != nil {
		return fmt.Errorf("apply cmd failed: %w", err)
	}

	return nil
}

func (m manager) Delete(ctx context.Context, resources []model.Resource) error {
	err := m.execute(ctx, resources, m.deleteArgs)
	if err != nil {
		return fmt.Errorf("delete cmd failed: %w", err)
	}

	return nil
}

func (m manager) execute(ctx context.Context, resources []model.Resource, cmdArgs []string) error {
	if len(resources) == 0 {
		return nil
	}

	logger := m.logger.WithValues(log.Kv{"ext-cmd": "kubectl"})

	objs := make([]model.K8sObject, 0, len(resources))
	for _, r := range resources {
		objs = append(objs, r.K8sObject)
	}

	// Encode objects.
	yamlData, err := m.yamlEncoder.EncodeObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("could not encode objects: %w", err)
	}

	// Create command.
	var errOut bytes.Buffer
	in := bytes.NewReader(yamlData)
	cmd := exec.CommandContext(ctx, m.kubectlCmd, cmdArgs...)
	cmd.Stdin = in
	cmd.Stderr = &errOut
	stdoutStream, err := m.cmdRunner.StdoutPipe(cmd)
	if err != nil {
		return fmt.Errorf("error while streaming stdout on command: %w", err)
	}

	// Execute command command in streaming mode with our logger.
	err = m.cmdRunner.Start(cmd)
	if err != nil {
		return fmt.Errorf("error while starting command: %w", err)
	}

	// Stream out.
	stream := bufio.NewScanner(stdoutStream)
	for stream.Scan() {
		logger.Infof(stream.Text())
	}

	err = m.cmdRunner.Wait(cmd)
	if err != nil {
		stderrData := errOut.String()
		for _, line := range strings.Split(stderrData, "\n") {
			if line == "" {
				continue
			}
			logger.Errorf(line)
		}
		return fmt.Errorf("error on cmd execution: %s: %w", stderrData, err)
	}

	return nil
}
