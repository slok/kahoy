package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// ObjectLoader knows how to load Kubernetes object from manifests.
type ObjectLoader interface {
	LoadObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error)
}

type yamlObjectLoader struct {
	serializer runtime.Serializer
	logger     log.Logger
}

// NewYAMLObjectLoader returns a new YAML Kubernetes object loader.
func NewYAMLObjectLoader(logger log.Logger) ObjectLoader {
	return yamlObjectLoader{
		// Create a unstructured yaml decoder (we don't know what type of objects are we loading).
		serializer: yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),

		logger: logger.WithValues(log.Kv{"app-svc": "kubernetes.YAMLObjectLoader"}),
	}
}

var splitMarkRe = regexp.MustCompile("(?m)^---")

const emptyChars = "\n\t\r "

func (y yamlObjectLoader) LoadObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error) {
	// Santize and split (YAML can declar multiple files in the same file using `---`).
	raw = bytes.Trim(raw, emptyChars)
	rawSplit := splitMarkRe.Split(string(raw), -1)

	// Decode all objects in the raw.
	res := make([]model.K8sObject, 0, len(rawSplit))
	for _, rawObj := range rawSplit {
		// Sanitize and ignore if no data.
		rawObj = strings.Trim(rawObj, emptyChars)
		if rawObj == "" {
			continue
		}

		obj, _, err := y.serializer.Decode([]byte(rawObj), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("could not decode kubernetes object %w", err)
		}
		res = append(res, obj.(*unstructured.Unstructured))
	}

	return res, nil
}
