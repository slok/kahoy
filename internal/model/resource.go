package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// K8sObject is the k8s object itself.
type K8sObject interface {
	runtime.Object
	metav1.Object
}

// ResourceState represents the state of a resource.
type ResourceState string

const (
	// ResourceStateExists represents a state where the resource should exists.
	ResourceStateExists ResourceState = "exists"
	// ResourceStateMissing represents a state where the resource should be missing.
	ResourceStateMissing ResourceState = "missing"
)

// Resource representes a resource.
type Resource struct {
	ID    string
	Name  string
	State ResourceState
}

// ResourceGroup is a group or resources.
type ResourceGroup struct {
	ID        string
	Name      string
	Resources []Resource
}
