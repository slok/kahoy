package kubectl

import (
	"context"
	"io"
	"os/exec"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// CmdRunner knows how to run exec.Cmd commands. Normally we use this
// so we can mock the cmd run without the need to wait for the streams
// capture the command executed...
type CmdRunner interface {
	Run(c *exec.Cmd) error
	Start(c *exec.Cmd) error
	Wait(c *exec.Cmd) error
	StdoutPipe(c *exec.Cmd) (io.ReadCloser, error)
}

//go:generate mockery --case underscore --output kubectlmock --outpkg kubectlmock --name CmdRunner

type stdCmdRunner struct {
	logger log.Logger
}

func newStdCmdRunner(logger log.Logger) CmdRunner {
	return stdCmdRunner{
		logger: logger.WithValues(log.Kv{"app-svc": "kubectl.stdCmdRunner"}),
	}
}

func (s stdCmdRunner) Run(c *exec.Cmd) error {
	s.logger.Debugf("running cmd: '%s'", c)
	return c.Run()
}

func (s stdCmdRunner) Start(c *exec.Cmd) error {
	s.logger.Debugf("running cmd: '%s'", c)
	return c.Start()
}

func (s stdCmdRunner) Wait(c *exec.Cmd) error {
	return c.Wait()
}

func (s stdCmdRunner) StdoutPipe(c *exec.Cmd) (io.ReadCloser, error) {
	return c.StdoutPipe()
}

// K8sObjectEncoder knows how to encode K8s objects into Raw Kubernetes compatible formats.
type K8sObjectEncoder interface {
	EncodeObjects(ctx context.Context, objs []model.K8sObject) ([]byte, error)
}

//go:generate mockery --case underscore --output kubectlmock --outpkg kubectlmock --name K8sObjectEncoder

// K8sObjectDecoder knows how to decode raw YAML Kubernetes into models.
type K8sObjectDecoder interface {
	DecodeObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error)
}

//go:generate mockery --case underscore --output kubectlmock --outpkg kubectlmock --name K8sObjectDecoder
