package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource"
)

// FSManager knows how to manage resources on the fs.
type FSManager interface {
	TempDir(dir, pattern string) (name string, err error)
	RemoveAll(path string) error
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type stdFSManager struct{}

//go:generate mockery --case underscore --output kubectlmock --outpkg kubectlmock --name FSManager

func (stdFSManager) TempDir(dir, pattern string) (name string, err error) {
	return ioutil.TempDir(dir, pattern)
}

func (stdFSManager) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (stdFSManager) RemoveAll(path string) error { return os.RemoveAll(path) }

// DiffManagerConfig is the configuration for NewDiffManager.
type DiffManagerConfig struct {
	KubectlCmd                string
	KubeConfig                string
	KubeContext               string
	KubeFieldManager          string
	DisableKubeForceConflicts bool
	YAMLEncoder               K8sObjectEncoder
	FSManager                 FSManager
	CmdRunner                 CmdRunner
	Out                       io.Writer
	ErrOut                    io.Writer
	Logger                    log.Logger
}

func (c *DiffManagerConfig) defaults() error {
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
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "kubectl.DiffManager"})

	if c.CmdRunner == nil {
		c.CmdRunner = stdCmdRunner{
			logger: c.Logger.WithValues(log.Kv{"app-svc": "kubectl.StdCmdRunner"}),
		}
	}

	if c.YAMLEncoder == nil {
		return fmt.Errorf("yaml encoder is required")
	}

	if c.FSManager == nil {
		c.FSManager = stdFSManager{}
	}

	return nil
}

type diffManager struct {
	kubectlCmd  string
	yamlEncoder K8sObjectEncoder
	cmdRunner   CmdRunner
	fsManager   FSManager
	out         io.Writer
	errOut      io.Writer
	logger      log.Logger

	applyArgs []string
}

// NewDiffManager returns a resource Manager based on Kubctl that will
// output diff changes.
func NewDiffManager(config DiffManagerConfig) (resource.Manager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	applyArgs := newKubectlCmdArgs([]kubectlCmdOption{
		withDiffCmd(),
		withContext(config.KubeContext),
		withConfig(config.KubeConfig),
		withForceConflicts(!config.DisableKubeForceConflicts),
		withFieldManager(config.KubeFieldManager),
		withServerSide(true),
		withStdIn(),
	})

	return diffManager{
		kubectlCmd:  config.KubectlCmd,
		yamlEncoder: config.YAMLEncoder,
		cmdRunner:   config.CmdRunner,
		fsManager:   config.FSManager,
		out:         config.Out,
		errOut:      config.ErrOut,
		logger:      config.Logger,
		applyArgs:   applyArgs,
	}, nil
}

func (d diffManager) Apply(ctx context.Context, resources []model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	objs := make([]model.K8sObject, 0, len(resources))
	for _, r := range resources {
		objs = append(objs, r.K8sObject)
	}

	// Encode objects.
	yamlData, err := d.yamlEncoder.EncodeObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("could not encode objects to diff: %w", err)
	}

	// Create command.
	in := bytes.NewReader(yamlData)
	cmd := d.newKubctlCmd(ctx, d.applyArgs, in)

	// Execute command.
	err = d.cmdRunner.Run(cmd)
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		// No error if our error is 1 exit code, just changes on diff.
		// Check: https://github.com/kubernetes/kubernetes/pull/87437
		if ok && exitErr.ExitCode() < 2 {
			return nil
		}

		return fmt.Errorf("error while running apply diff command: %w", err)
	}

	return nil
}

// TODO(slok): At this moment we are making client-side diff with the stored state instead the server one.
// 			   Implement server-side diff, get the resource and check what needs to be removed.
func (d diffManager) Delete(ctx context.Context, resources []model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	// Create a temporal directory.
	dirPath, err := d.fsManager.TempDir("", "KAHOY-")
	if err != nil {
		return fmt.Errorf("could not create temp dir: %w", err)
	}
	defer func() {
		err := d.fsManager.RemoveAll(dirPath)
		if err != nil {
			d.logger.Warningf("could not delete %q diff temp dir: %s", dirPath, err)
		}
	}()

	// We are creating a diff for each resource.
	for _, r := range resources {
		// Encode object.
		yamlData, err := d.yamlEncoder.EncodeObjects(ctx, []model.K8sObject{r.K8sObject})
		if err != nil {
			return fmt.Errorf("could not encode objects to diff: %w", err)
		}

		// Store our YAML in a file.
		gvk := r.K8sObject.GetObjectKind().GroupVersionKind()
		fileName := strings.Join([]string{gvk.Group, gvk.Version, gvk.Kind, r.K8sObject.GetNamespace(), r.K8sObject.GetName()}, ".")
		filePath := path.Join(dirPath, fileName)
		err = d.fsManager.WriteFile(filePath, yamlData, 0644)
		if err != nil {
			return fmt.Errorf("could not write in file: %w", err)
		}

		// Create a diff commmand with the 2nd file as empty and execute.
		cmd := d.newDiffCmd(ctx, filePath, strings.NewReader(""))
		err = d.cmdRunner.Run(cmd)
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			// No error if our error is 1 exit code, just changes on diff.
			if ok && exitErr.ExitCode() >= 2 {
				return fmt.Errorf("error while running delete diff command: %w", err)
			}
		}
	}

	return nil
}

func (d diffManager) newDiffCmd(ctx context.Context, fileA string, in io.Reader) *exec.Cmd {
	args := []string{"-u", "-N", fileA, "-"}
	cmd := exec.CommandContext(ctx, "diff", args...)
	cmd.Stdin = in
	cmd.Stdout = d.out
	cmd.Stderr = d.errOut
	return cmd
}

func (d diffManager) newKubctlCmd(ctx context.Context, args []string, in io.Reader) *exec.Cmd {
	cmd := exec.CommandContext(ctx, d.kubectlCmd, args...)
	cmd.Stdin = in
	cmd.Stdout = d.out
	cmd.Stderr = d.errOut
	return cmd
}
