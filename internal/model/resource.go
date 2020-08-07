package model

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// K8sObject is the k8s object itself.
type K8sObject interface {
	runtime.Object
}

// Resource representes a resource.
type Resource struct {
	ID           string
	GroupID      string
	Name         string
	ManifestPath string
	K8sObject    K8sObject
}

// Group represents a group of resources.
type Group struct {
	ID string
}
