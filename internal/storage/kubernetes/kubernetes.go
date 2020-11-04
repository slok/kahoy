package kubernetes

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
)

// K8sClient knows how to interact with a Kubernetes apiserver.
type K8sClient interface {
	EnsureSecret(ctx context.Context, sec *corev1.Secret) error
	EnsureMissingSecret(ctx context.Context, ns, name string) error
	ListSecrets(ctx context.Context, ns string, labelFilter map[string]string) ([]corev1.Secret, error)
	GetSecret(ctx context.Context, ns, name string) (*corev1.Secret, error)
}

//go:generate mockery --case underscore --output kubernetesmock --outpkg kubernetesmock --name K8sClient

// K8sObjectSerializer knows how to decode/encode K8s objects into text based raw formats.
type K8sObjectSerializer interface {
	EncodeObjects(ctx context.Context, objs []model.K8sObject) ([]byte, error)
	DecodeObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error)
}

//go:generate mockery --case underscore --output kubernetesmock --outpkg kubernetesmock --name K8sObjectSerializer

// RepositoryConfig is the configuration of the Repository.
type RepositoryConfig struct {
	// Namespace is the namespace where Kahoy will store the state.
	Namespace string
	// StorageID is the id that identifies the state stored. This is important
	// because kahoy can be run N times with different params (e.g manifest paths)
	// and those would be two different state stores. StorageID is what Kahoy uses
	// to keep those independent.
	//
	// StorageID has the same requirements as a Kubernetes label value.
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
	StorageID    string
	Serializer   K8sObjectSerializer
	Client       K8sClient
	ModelFactory *model.ResourceAndGroupFactory
	Logger       log.Logger
}

func (c *RepositoryConfig) defaults() error {
	if c.Namespace == "" {
		c.Namespace = "default"
	}

	if c.StorageID == "" {
		return fmt.Errorf("storage ID is required")
	}

	// Validate storage ID.
	errStrs := validation.IsValidLabelValue(c.StorageID)
	if len(errStrs) > 0 {
		return fmt.Errorf("invalid storageID: %s", strings.Join(errStrs, ":"))
	}

	if c.Serializer == nil {
		return fmt.Errorf("serializer is required")
	}

	if c.Client == nil {
		return fmt.Errorf("kubernetes client is required")
	}

	if c.ModelFactory == nil {
		return fmt.Errorf("resource and group model factory is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "kubernetes.Repository"})

	return nil
}

// Repository knows how to store and load resources from a K8s storage (apiserver).
type Repository struct {
	namespace    string
	storageID    string
	serializer   K8sObjectSerializer
	client       K8sClient
	modelFactory *model.ResourceAndGroupFactory
}

var (
	_ storage.StateRepository    = Repository{}
	_ storage.ResourceRepository = Repository{}
)

// NewRepository returns a new respoitory.
func NewRepository(config RepositoryConfig) (*Repository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Repository{
		namespace:    config.Namespace,
		storageID:    config.StorageID,
		serializer:   config.Serializer,
		client:       config.Client,
		modelFactory: config.ModelFactory,
	}, nil
}

// GetResource satisfies storage.ResourceRepository interface.
func (r Repository) GetResource(ctx context.Context, id string) (*model.Resource, error) {
	secret, err := r.client.GetSecret(ctx, r.namespace, r.genK8sName(id))
	if err != nil {
		return nil, fmt.Errorf("could not get secret from Kubernetes: %w", err)
	}

	res, err := r.secretToResource(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("could not convert %q secret to model resource: %w", secret.GetName(), err)
	}

	return res, nil
}

// ListResources satisfies storage.ResourceRepository interface.
func (r Repository) ListResources(ctx context.Context, opts storage.ResourceListOpts) (*storage.ResourceList, error) {
	secretList, err := r.client.ListSecrets(ctx, r.namespace, r.genK8sLabels())
	if err != nil {
		return nil, fmt.Errorf("could not list secrets from Kubernetes: %w", err)
	}

	resList := storage.ResourceList{}
	for _, secret := range secretList {
		res, err := r.secretToResource(ctx, &secret)
		if err != nil {
			return nil, fmt.Errorf("could not convert %q secret to model resource: %w", secret.GetName(), err)
		}
		resList.Items = append(resList.Items, *res)
	}

	return &resList, nil
}

// StoreState satisfies storage.StateRepository interface.
func (r Repository) StoreState(ctx context.Context, state model.State) error {
	// Store applied resources.
	for _, res := range state.AppliedResources {
		err := r.storeResource(ctx, res)
		if err != nil {
			return fmt.Errorf("could not store applied resource: %w", err)
		}
	}

	// Delete deleted resources.
	for _, res := range state.DeletedResources {
		err := r.deleteResource(ctx, res)
		if err != nil {
			return fmt.Errorf("could not store applied resource: %w", err)
		}
	}

	return nil
}

const kubePathFmt = "kubernetes://%s/%s"

func (r Repository) secretToResource(ctx context.Context, secret *corev1.Secret) (*model.Resource, error) {
	_, ok := secret.Data[secretResIDKey]
	if !ok {
		return nil, fmt.Errorf("missing resource id in kubernetes secret")
	}

	group, ok := secret.Data[secretResGroupKey]
	if !ok {
		return nil, fmt.Errorf("missing resource group in kubernetes secret")
	}

	raw, ok := secret.Data[secretResDataKey]
	if !ok {
		return nil, fmt.Errorf("missing resource data in kubernetes secret")
	}

	_, ok = secret.Data[secretResPathKey]
	if !ok {
		return nil, fmt.Errorf("missing resource fs path in kubernetes secret")
	}

	kobj, err := r.deserialize(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("could not deserialize kubernetes object data: %w", err)
	}

	path := fmt.Sprintf(kubePathFmt, secret.Namespace, secret.Name)

	return r.modelFactory.NewResource(kobj, string(group), path)
}

const (
	secretResDataKey  = "raw"
	secretResIDKey    = "id"
	secretResGroupKey = "group"
	secretResPathKey  = "path"
)

func (r Repository) storeResource(ctx context.Context, res model.Resource) error {
	resourceData, err := r.serialize(ctx, res.K8sObject)
	if err != nil {
		return fmt.Errorf("could not serialize resource: %w", err)
	}

	resourceSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.genK8sName(res.ID),
			Namespace:   r.namespace,
			Labels:      r.genK8sLabels(),
			Annotations: r.genK8sAnnotations(res),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			secretResDataKey:  resourceData,
			secretResIDKey:    []byte(res.ID),
			secretResGroupKey: []byte(res.GroupID),
			secretResPathKey:  []byte(res.ManifestPath),
		},
	}

	err = r.client.EnsureSecret(ctx, resourceSecret)
	if err != nil {
		return fmt.Errorf("could not store secret on Kubernetes: %w", err)
	}

	return nil
}

func (r Repository) deleteResource(ctx context.Context, res model.Resource) error {
	err := r.client.EnsureMissingSecret(ctx, r.namespace, r.genK8sName(res.ID))
	if err != nil {
		return fmt.Errorf("could not delete secret on Kubernetes: %w", err)
	}

	return nil
}

func (r Repository) genK8sName(resID string) string {
	// Idempotent fixed length ID. StorageID is important to avoid collisions on same IDs with different runs.
	id := md5.Sum([]byte(fmt.Sprintf("kahoy.slok.dev-%s-%s", r.storageID, resID)))
	return fmt.Sprintf("%x", id)
}

func (r Repository) genK8sLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "kahoy",
		"app.kubernetes.io/component":  "internal",
		"app.kubernetes.io/part-of":    "storage",
		"app.kubernetes.io/managed-by": "kahoy",
		"kahoy.slok.dev/storage-id":    r.storageID,
	}
}

func (r Repository) genK8sAnnotations(res model.Resource) map[string]string {
	return map[string]string{
		"kahoy.slok.dev/resource-id":    res.ID,
		"kahoy.slok.dev/resource-group": res.GroupID,
		"kahoy.slok.dev/resource-name":  res.K8sObject.GetName(),
		"kahoy.slok.dev/resource-ns":    res.K8sObject.GetNamespace(),
	}
}

// serialize serializes the model.Resource data so is able to store.
// To serialize we are going to:
// - Serialize our Kubernetes models to text based (e.g YAML).
// - Compress with gz.
func (r Repository) serialize(ctx context.Context, obj model.K8sObject) ([]byte, error) {
	// Encode data.
	data, err := r.serializer.EncodeObjects(ctx, []model.K8sObject{obj})
	if err != nil {
		return nil, fmt.Errorf("could not serialize object: %w", err)
	}

	// Compress data.
	var b bytes.Buffer
	gzw := gzip.NewWriter(&b)
	_, err = gzw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("could not compress serialized resource data: %w", err)
	}
	err = gzw.Close()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(&b)
}

// deserialize does the opposite to serialize:
// - Decompress with gz.
// - Deserialize our Kubernetes raw data to model.
func (r Repository) deserialize(ctx context.Context, data []byte) (model.K8sObject, error) {
	// Decompress data.
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not load compressed data: %w", err)
	}
	decData, err := ioutil.ReadAll(gzr)
	if err != nil {
		return nil, fmt.Errorf("could not decompress data: %w", err)
	}

	// Decode data.
	objs, err := r.serializer.DecodeObjects(ctx, decData)
	if err != nil {
		return nil, fmt.Errorf("decode kubernetes object data: %w", err)
	}

	if len(objs) != 1 {
		return nil, fmt.Errorf("wrong number of decoded kubernetes objects: %w", err)
	}

	return objs[0], nil
}
