package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/slok/kahoy/internal/log"
)

// K8sObject is the k8s object itself.
type K8sObject interface {
	runtime.Object
	metav1.Object
}

// Resource represents a resource.
type Resource struct {
	ID           string
	GroupID      string
	ManifestPath string
	K8sObject    K8sObject
}

// Group represents a group of resources.
type Group struct {
	ID       string
	Path     string
	Priority int
	Hooks    GroupHooks
}

// GroupHooks tells what are the hooks.
type GroupHooks struct {
	Pre  *GroupHookSpec
	Post *GroupHookSpec
}

// GroupHookSpec are the hook options.
type GroupHookSpec struct {
	Cmd     string
	Timeout time.Duration
}

// KubernetesDiscoveryClient is the client used to discover resource types on
// a Kubernetes cluster.
type KubernetesDiscoveryClient interface {
	GetServerGroupsAndResources(ctx context.Context) ([]*metav1.APIGroup, []*metav1.APIResourceList, error)
}

//go:generate mockery --case underscore --output modelmock --outpkg modelmock --name KubernetesDiscoveryClient

// ResourceAndGroupFactory knows how to return new resource and group models.
type ResourceAndGroupFactory struct {
	kubeAPITypesCache map[string]metav1.APIResource
	logger            log.Logger
}

// NewResourceAndGroupFactory returns a new ResourceAndGroupFactory.
func NewResourceAndGroupFactory(cli KubernetesDiscoveryClient, logger log.Logger) (*ResourceAndGroupFactory, error) {
	// Get groups and resources from the apiserver.
	_, res, err := cli.GetServerGroupsAndResources(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes API resources: %w", err)
	}

	// Index Kubernetes types information received from the cluster.
	kubeAPITypes := map[string]metav1.APIResource{}
	for _, re := range res {
		gv := re.GroupVersion
		for _, r := range re.APIResources {
			id := strings.Trim(fmt.Sprintf("%s/%s", gv, r.Kind), "/")
			kubeAPITypes[id] = r
		}
	}

	return &ResourceAndGroupFactory{
		kubeAPITypesCache: kubeAPITypes,
		logger:            logger.WithValues(log.Kv{"app-svc": "model.ResourceAndGroupFactory"}),
	}, nil
}

// NewResource creates a new initialized resource.
// It will ensure the ID generated for the resource is correct and the resources is a correct type.
func (r ResourceAndGroupFactory) NewResource(k8sObject K8sObject, groupID string, manifestPath string) (*Resource, error) {
	gvk := k8sObject.GetObjectKind().GroupVersionKind()
	id := strings.Trim(fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind), "/")

	resType, ok := r.kubeAPITypesCache[id]
	if !ok {
		// TODO(slok): Should we allow loading unknown types by the server?.
		return nil, fmt.Errorf("unknown Kubernetes resource type by the apiserver: %s", id)
	}

	modelID := r.genResourceID(k8sObject)
	// If k8s object is cluster scoped and has a namespace set we need to create the ID
	// again because the cluster scoped resources should have always `default` as the namespace
	// in the ID part (tl;dr: cluster scoped ignore namespace field and use always `default` ns).
	if !resType.Namespaced && k8sObject.GetNamespace() != "" {
		r.logger.WithValues(log.Kv{
			"resource-id":       modelID,
			"resource-group-id": groupID,
			"resource-path":     manifestPath,
			"kube-type":         id,
		}).Warningf("cluster scoped resource has namespace set")
		tmp := k8sObject.DeepCopyObject()
		k8sObjectCopy := tmp.(K8sObject)
		k8sObjectCopy.SetNamespace("")
		modelID = r.genResourceID(k8sObjectCopy)
	}

	return &Resource{
		ID:           modelID,
		GroupID:      groupID,
		ManifestPath: manifestPath,
		K8sObject:    k8sObject,
	}, nil
}

func (r ResourceAndGroupFactory) genResourceID(obj K8sObject) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	group := "core"
	if gvk.Group != "" {
		group = gvk.Group
	}
	ns := "default"
	if obj.GetNamespace() != "" {
		ns = obj.GetNamespace()
	}

	return fmt.Sprintf("%s/%s/%s/%s/%s", group, gvk.Version, gvk.Kind, ns, obj.GetName())
}

// NewGroup creates a new group.
func (r ResourceAndGroupFactory) NewGroup(id, path string, config GroupConfig) Group {
	const defaultPriority = 1000

	g := Group{
		ID:   id,
		Path: path,
	}

	// Set priority.
	g.Priority = defaultPriority
	if config.Priority != nil {
		g.Priority = *config.Priority
	}

	// Set wait options.
	if config.HooksConfig.Pre != nil {
		g.Hooks.Pre = waitConfigToGroupModel(*config.HooksConfig.Pre)
	}
	if config.HooksConfig.Post != nil {
		g.Hooks.Post = waitConfigToGroupModel(*config.HooksConfig.Post)
	}

	return g
}

func waitConfigToGroupModel(cfg GroupHookConfigSpec) *GroupHookSpec {
	return &GroupHookSpec{
		Cmd:     cfg.Cmd,
		Timeout: cfg.Timeout,
	}
}
