package model

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// K8sObject is the k8s object itself.
type K8sObject interface {
	runtime.Object
	metav1.Object
}

// Resource representes a resource.
type Resource struct {
	ID           string
	GroupID      string
	Name         string
	ManifestPath string
	K8sObject    K8sObject
}

// GenResourceID knows how to get a resource ID.
func GenResourceID(obj K8sObject) string {
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

// Group represents a group of resources.
type Group struct {
	ID       string
	Path     string
	Priority int
}

// NewGroup returns a new group using app group configuration.
func NewGroup(id, path string, config GroupConfig) Group {
	const defaultPriority = 1000

	priority := defaultPriority
	if config.Priority != nil {
		priority = *config.Priority
	}

	return Group{
		ID:       id,
		Path:     path,
		Priority: priority,
	}
}
