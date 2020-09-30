package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	envKahoyBin         = "KAHOY_INTEGRATION_BINARY"
	envKahoyKubeContext = "KAHOY_INTEGRATION_KUBE_CONTEXT"
	envKahoyKubeConfig  = "KAHOY_INTEGRATION_KUBE_CONFIG"
)

// Config is the configuration for integration tests.
type Config struct {
	Binary      string
	KubeConfig  string
	KubeContext string
}

func (c *Config) defaults() error {
	if c.Binary == "" {
		c.Binary = "kahoy"
	}

	_, err := exec.LookPath(c.Binary)
	if err != nil {
		return fmt.Errorf("kahoy binary missing in %q: %w", c.Binary, err)
	}

	return nil
}

// GetIntegrationConfig returns the configuration of the
func GetIntegrationConfig(ctx context.Context) (*Config, error) {
	c := &Config{
		Binary:      os.Getenv(envKahoyBin),
		KubeConfig:  os.Getenv(envKahoyKubeContext),
		KubeContext: os.Getenv(envKahoyKubeConfig),
	}

	err := c.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return c, nil
}

// RunKahoy executes kahoy command.
func RunKahoy(ctx context.Context, config Config, kahoyCmd string) (stdout, stderr []byte, err error) {
	// Sanitize command.
	kahoyCmd = strings.TrimSpace(kahoyCmd)
	args := strings.Split(kahoyCmd, " ")
	if args[0] == "kahoy" || args[0] == config.Binary {
		args = args[1:]
	}

	// Create command.
	var outData, errData bytes.Buffer
	cmd := exec.CommandContext(ctx, config.Binary, args...)
	cmd.Stdout = &outData
	cmd.Stderr = &errData

	// Set env.
	newenv := []string{
		"KAHOY_AUTO_APPROVE=true",
		"KAHOY_NO_LOG=true",
		"KAHOY_NO_COLOR=true",
	}
	if config.KubeConfig != "" {
		newenv = append(newenv, fmt.Sprintf("KAHOY_KUBE_CONFIG=%s", config.KubeConfig))
	}
	if config.KubeContext != "" {
		newenv = append(newenv, fmt.Sprintf("KAHOY_KUBE_CONTEXT=%s", config.KubeContext))
	}
	cmd.Env = append(os.Environ(), newenv...)

	// Run.
	err = cmd.Run()

	return outData.Bytes(), errData.Bytes(), err
}
