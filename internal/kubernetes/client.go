package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/slok/kahoy/internal/log"
	storagekubernetes "github.com/slok/kahoy/internal/storage/kubernetes"
)

// Client knows how to interact with Kubernetes apiserver. Is a small layer on top
// of regular Kubernetes client.
type Client struct {
	coreCli kubernetes.Interface
	logger  log.Logger
}

var _ storagekubernetes.K8sClient = Client{}

// NewClient returns a new Kubernetes client.
func NewClient(coreCli kubernetes.Interface, logger log.Logger) Client {
	return Client{
		coreCli: coreCli,
		logger:  logger.WithValues(log.Kv{"app-svc": "kubernetes.Client"}),
	}
}

// GetSecret gets a secret from Kubernetes.
func (c Client) GetSecret(ctx context.Context, ns, name string) (*corev1.Secret, error) {
	logger := c.logger.WithValues(log.Kv{"obj-ns": ns, "obj-name": name})

	secret, err := c.coreCli.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	logger.Debugf("secret retrieved")

	return secret, nil
}

// ListSecrets lists secrts.
func (c Client) ListSecrets(ctx context.Context, ns string, labelFilter map[string]string) ([]corev1.Secret, error) {
	logger := c.logger.WithValues(log.Kv{"obj-ns": ns})

	opts := metav1.ListOptions{
		LabelSelector: labels.Set(labelFilter).String(),
	}

	secrets := []corev1.Secret{}
	for {
		secretList, err := c.coreCli.CoreV1().Secrets(ns).List(ctx, opts)
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, secretList.Items...)

		// Check if we have more objects.
		if secretList.Continue == "" {
			break
		}
		opts.Continue = secretList.Continue
	}

	logger.Debugf("%d secrets retrieved", len(secrets))

	return secrets, nil
}

// EnsureSecret creates the secret if not present, and overrides if present.
// TODO(slok): Check server-side apply.
func (c Client) EnsureSecret(ctx context.Context, secret *corev1.Secret) error {
	logger := c.logger.WithValues(log.Kv{"obj-ns": secret.Namespace, "obj-name": secret.Name})

	storedSecret, err := c.coreCli.CoreV1().Secrets(secret.Namespace).Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			return err
		}
		_, err = c.coreCli.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logger.Debugf("secret has been created")

		return nil
	}

	// Force overwrite.
	secret.ObjectMeta.ResourceVersion = storedSecret.ResourceVersion
	_, err = c.coreCli.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	logger.Debugf("secret has been updated")

	return nil
}

// EnsureMissingSecret will delete the secret if exists, and noop if doesn't exists.
func (c Client) EnsureMissingSecret(ctx context.Context, ns, name string) error {
	logger := c.logger.WithValues(log.Kv{"obj-ns": ns, "obj-name": name})

	err := c.coreCli.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !kubeerrors.IsNotFound(err) {
		return err
	}

	logger.Debugf("secret has been deleted")
	return nil
}
