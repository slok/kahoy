package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage/kubectl"
	storagefs "github.com/slok/kahoy/internal/storage/fs"
)

// YAMLObjectSerializer handles YAML based raw data, by decoding and encoding from/into
// Kubernetes model objects.
type YAMLObjectSerializer struct {
	encoder runtime.Encoder
	decoder runtime.Decoder
	logger  log.Logger
}

// Interface assertion.
var (
	_ storagefs.K8sObjectDecoder = YAMLObjectSerializer{}
	_ kubectl.K8sObjectEncoder   = YAMLObjectSerializer{}
)

// NewYAMLObjectSerializer returns a new YAMLNewYAMLObjectSerializer.
func NewYAMLObjectSerializer(logger log.Logger) YAMLObjectSerializer {
	return YAMLObjectSerializer{
		// Create a unstructured yaml decoder (we don't know what type of objects are we loading).
		encoder: json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil),
		decoder: yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		logger:  logger.WithValues(log.Kv{"app-svc": "kubernetes.YAMLObjectSerializer"}),
	}
}

var (
	splitMarkRe  = regexp.MustCompile("(?m)^---")
	rmCommentsRe = regexp.MustCompile("(?m)^#.*$")
)

const emptyChars = "\n\t\r "

// DecodeObjects decodes YAML data into objects, supports multiple objects on the same
// YAML raw data.
func (y YAMLObjectSerializer) DecodeObjects(ctx context.Context, raw []byte) ([]model.K8sObject, error) {
	// Santize and split (YAML can declare multiple files in the same file using `---`).
	raw = bytes.Trim(raw, emptyChars)
	raw = rmCommentsRe.ReplaceAll(raw, []byte(""))
	rawSplit := splitMarkRe.Split(string(raw), -1)

	// Decode all objects in the raw.
	res := make([]model.K8sObject, 0, len(rawSplit))
	for _, rawObj := range rawSplit {
		// Sanitize and ignore if no data.
		rawObj = strings.Trim(rawObj, emptyChars)
		if rawObj == "" {
			continue
		}

		obj, _, err := y.decoder.Decode([]byte(rawObj), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("could not decode kubernetes object %w", err)
		}

		switch objt := obj.(type) {
		case *unstructured.Unstructured:
			res = append(res, objt)
		case *unstructured.UnstructuredList:
			// If a metav1.List type object, then get all the items individually
			// and add them to the resource list.
			for _, kobj := range objt.Items {
				kobj := kobj
				res = append(res, &kobj)
			}
		default:
			return nil, fmt.Errorf("decoded object is of an unknown type")
		}
	}

	return res, nil
}

// EncodeObjects encodes Kubernetes objects into YAML data, supports multiple objects on the same
// YAML raw data.
func (y YAMLObjectSerializer) EncodeObjects(ctx context.Context, objs []model.K8sObject) ([]byte, error) {
	var buffer bytes.Buffer
	for _, obj := range objs {
		if obj == nil {
			continue
		}

		_, _ = buffer.WriteString("---\n")

		err := y.encoder.Encode(obj, &buffer)
		if err != nil {
			return nil, fmt.Errorf("could not encode %s/%s kubernetes object: %w", obj.GetNamespace(), obj.GetName(), err)
		}
	}

	data, err := ioutil.ReadAll(&buffer)
	if err != nil {
		return nil, fmt.Errorf("could not read data from buffer: %w", err)
	}

	return data, nil
}
