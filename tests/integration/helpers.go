package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// IntegrationTestsNamespace is the namespace where the tests are being executed.
const IntegrationTestsNamespace = "kahoy-integration-test"

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

	if c.KubeConfig == "" {
		c.KubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	return nil
}

// GetIntegrationConfig returns the configuration of the integration tests environment.
func GetIntegrationConfig(ctx context.Context) (*Config, error) {
	const (
		envKahoyBin         = "KAHOY_INTEGRATION_BINARY"
		envKahoyKubeContext = "KAHOY_INTEGRATION_KUBE_CONTEXT"
		envKahoyKubeConfig  = "KAHOY_INTEGRATION_KUBE_CONFIG"
	)

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

// NewKubernetesClient returns a new client.
func NewKubernetesClient(ctx context.Context, config Config) (kubernetes.Interface, error) {
	kConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: config.KubeConfig,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: config.KubeContext,
			Timeout:        "3s",
		},
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("could not load Kubernetes configuration: %w", err)
	}

	cli, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create client-go kubernetes client: %w", err)
	}

	return cli, nil
}

// CleanTestsNamespace knows how to clean test namespace.
func CleanTestsNamespace(ctx context.Context, cli kubernetes.Interface) error {
	err := cli.CoreV1().Namespaces().Delete(context.TODO(), IntegrationTestsNamespace, metav1.DeleteOptions{})
	if err != nil && !kubeerrors.IsNotFound(err) {
		return err
	}

	// Wait.
	ticker := time.NewTicker(200 * time.Millisecond)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for namespace cleanup")
		}

		// Check if deleted.
		_, err := cli.CoreV1().Namespaces().Get(context.TODO(), IntegrationTestsNamespace, metav1.GetOptions{})
		if err != nil && kubeerrors.IsNotFound(err) {
			break
		}
	}

	return nil
}

var multiSpaceRegex = regexp.MustCompile(" +")

// RunKahoy executes kahoy command.
func RunKahoy(ctx context.Context, config Config, kahoyCmd string) (stdout, stderr []byte, err error) {
	// Sanitize command.
	kahoyCmd = strings.TrimSpace(kahoyCmd)
	kahoyCmd = multiSpaceRegex.ReplaceAllString(kahoyCmd, " ")

	// Split into args.
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
