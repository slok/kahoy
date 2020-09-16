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
	"github.com/slok/kahoy/internal/resource/manage"
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
	YAMLDecoder               K8sObjectDecoder
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
		c.CmdRunner = newStdCmdRunner(c.Logger)
	}

	if c.YAMLEncoder == nil {
		return fmt.Errorf("yaml encoder is required")
	}

	if c.YAMLDecoder == nil {
		return fmt.Errorf("yaml decoder is required")
	}

	if c.FSManager == nil {
		c.FSManager = stdFSManager{}
	}

	return nil
}

type diffManager struct {
	kubectlCmd  string
	yamlEncoder K8sObjectEncoder
	yamlDecoder K8sObjectDecoder
	cmdRunner   CmdRunner
	fsManager   FSManager
	out         io.Writer
	errOut      io.Writer
	logger      log.Logger

	applyArgs  []string
	deleteArgs []string
}

// NewDiffManager returns a resource Manager based on Kubctl that will
// output diff changes.
func NewDiffManager(config DiffManagerConfig) (manage.ResourceManager, error) {
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

	deleteArgs := newKubectlCmdArgs([]kubectlCmdOption{
		withGetCmd(),
		withContext(config.KubeContext),
		withConfig(config.KubeConfig),
		withIgnoreNotFound(true),
		withYAMLOutput(),
		withStdIn(),
	})

	return diffManager{
		kubectlCmd:  config.KubectlCmd,
		yamlEncoder: config.YAMLEncoder,
		yamlDecoder: config.YAMLDecoder,
		cmdRunner:   config.CmdRunner,
		fsManager:   config.FSManager,
		out:         config.Out,
		errOut:      config.ErrOut,
		logger:      config.Logger,
		applyArgs:   applyArgs,
		deleteArgs:  deleteArgs,
	}, nil
}

func (d diffManager) Apply(ctx context.Context, resources []model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	logger := d.logger.WithValues(log.Kv{"ext-cmd": "kubectl"})

	objs := make([]model.K8sObject, 0, len(resources))
	for _, r := range resources {
		objs = append(objs, r.K8sObject)
	}

	// Encode objects.
	yamlData, err := d.yamlEncoder.EncodeObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("could not encode objects to diff: %w", err)
	}

	// Create command. Diff output should go to stdout.
	in := bytes.NewReader(yamlData)
	var outErr bytes.Buffer
	cmd := exec.CommandContext(ctx, d.kubectlCmd, d.applyArgs...)
	cmd.Stdin = in
	cmd.Stdout = d.out
	cmd.Stderr = &outErr

	// Execute command.
	err = d.cmdRunner.Run(cmd)
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		// No error if our error is 1 exit code, just changes on diff.
		// Check: https://github.com/kubernetes/kubernetes/pull/87437
		if ok && exitErr.ExitCode() < 2 {
			return nil
		}

		for _, line := range strings.Split(outErr.String(), "\n") {
			if line == "" {
				continue
			}
			logger.Errorf(line)
		}

		return fmt.Errorf("error while running apply diff command: %s: %w", outErr.String(), err)
	}

	return nil
}

// Delete will get the diff for the deleted sources.
// We can't do the diff for things that will be deleted usin Kubectl, also we can't be sure of
// the resources that are on the server, maybe some of them don't exist neither on the server.
// To solve this, we ask the server for the current content of the resources that we want to delete,
// decode them to resources to validate individual resources and then get a diff against /dev/null
// for each object.
// With this solution, we get a real diff of that fields will be removed from the server.
// If anytime Kubectl handles deletion with diff, we should use that and remove all this logic.
func (d diffManager) Delete(ctx context.Context, resources []model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	// To get a real delete diff, we get the latest state that we have on the server
	// for the resources that we want to delete (changes, already deleted...).
	objs := make([]model.K8sObject, 0, len(resources))
	for _, r := range resources {
		objs = append(objs, r.K8sObject)
	}
	updatedObjs, err := d.getResourcesFromAPIServer(ctx, objs)
	if err != nil {
		return fmt.Errorf("could not get resources latest state from the apiserver: %w", err)
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
	logger := d.logger.WithValues(log.Kv{"ext-cmd": "diff"})
	for _, r := range updatedObjs {
		// Encode object.
		yamlData, err := d.yamlEncoder.EncodeObjects(ctx, []model.K8sObject{r})
		if err != nil {
			return fmt.Errorf("could not encode objects to diff: %w", err)
		}

		// Store our YAML in a file.
		gvk := r.GetObjectKind().GroupVersionKind()
		fileName := strings.Join([]string{gvk.Group, gvk.Version, gvk.Kind, r.GetNamespace(), r.GetName()}, ".")
		filePath := path.Join(dirPath, fileName)
		err = d.fsManager.WriteFile(filePath, yamlData, 0644)
		if err != nil {
			return fmt.Errorf("could not write in file: %w", err)
		}

		// Create a diff commmand with the 2nd file as empty and execute.
		var errOut bytes.Buffer
		args := []string{"-u", "-N", filePath, "-"}
		cmd := exec.CommandContext(ctx, "diff", args...)
		cmd.Stdin = strings.NewReader("")
		cmd.Stdout = d.out
		cmd.Stderr = &errOut

		err = d.cmdRunner.Run(cmd)
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			// No error if our error is 1 exit code, just changes on diff.
			if ok && exitErr.ExitCode() < 2 {
				continue
			}

			stderrData := errOut.String()
			for _, line := range strings.Split(stderrData, "\n") {
				if line == "" {
					continue
				}
				logger.Errorf(line)
			}

			return fmt.Errorf("error while running delete diff command: %s: %w", stderrData, err)
		}
	}

	return nil
}

// getResourcesFromAPIServer returns the latest state from apiserver of the received resources.
// Tries getting the latest information of the received resources by asking them to the server
// using kubectl, and decoding again into domain k8s objects.
// Existing resources will be returned with the fields data up to date (server data).
// If any of the resources is not on the server it will not be returned.
func (d diffManager) getResourcesFromAPIServer(ctx context.Context, objs []model.K8sObject) ([]model.K8sObject, error) {
	logger := d.logger.WithValues(log.Kv{"ext-cmd": "kubectl"})

	// Encode objects.
	yamlData, err := d.yamlEncoder.EncodeObjects(ctx, objs)
	if err != nil {
		return nil, fmt.Errorf("could not encode objects to diff: %w", err)
	}

	// Create command.
	var out, outErr bytes.Buffer
	in := bytes.NewReader(yamlData)
	cmd := exec.CommandContext(ctx, d.kubectlCmd, d.deleteArgs...)
	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = &outErr

	// Execute command.
	err = d.cmdRunner.Run(cmd)
	if err != nil {
		for _, line := range strings.Split(outErr.String(), "\n") {
			if line == "" {
				continue
			}
			logger.Errorf(line)
		}
		return nil, fmt.Errorf("error while running delete diff command: %s: %w", outErr.String(), err)
	}

	// Decode into model.
	k8sResources, err := d.yamlDecoder.DecodeObjects(ctx, out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("could not decode server obtained Kubernetes objects for deletion: %w", err)
	}

	return k8sResources, nil
}
