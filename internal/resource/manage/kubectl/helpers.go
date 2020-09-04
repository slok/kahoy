package kubectl

import (
	"context"
	"os/exec"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// CmdRunner knows how to run exec.Cmd commands.
type CmdRunner interface {
	Run(*exec.Cmd) error
}

//go:generate mockery --case underscore --output kubectlmock --outpkg kubectlmock --name CmdRunner

type stdCmdRunner struct {
	logger log.Logger
}

func (s stdCmdRunner) Run(c *exec.Cmd) error {
	s.logger.Debugf("running cmd: '%s'", c)
	return c.Run()
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
