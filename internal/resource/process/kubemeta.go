package process

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

// NewExcludeKubeTypeProcessor returns a new Resource processor that will exclude (remove)
// from the received resources, the ones that regex match the group version kind
// (e.g core/v1/pod) on the kubernetes resource.
func NewExcludeKubeTypeProcessor(kubeTypeRegex []string, logger log.Logger) (ResourceProcessor, error) {
	logger = logger.WithValues(log.Kv{"app-svc": "process.ExcludeKubeTypeProcessor"})

	compiledRegex := make([]*regexp.Regexp, 0, len(kubeTypeRegex))
	for _, r := range kubeTypeRegex {
		if r == "" {
			continue
		}
		cr, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("invalid regex %q: %w", r, err)
		}
		compiledRegex = append(compiledRegex, cr)
	}

	return ResourceProcessorFunc(func(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
		newRes := []model.Resource{}

		for _, r := range resources {
			gvk := r.K8sObject.GetObjectKind().GroupVersionKind()

			// Get `extensions/v1beta1/Ingress` or `v1/Pod` style kubernetes type.
			parts := []string{}
			if gvk.Group != "" {
				parts = append(parts, gvk.Group)
			}
			parts = append(parts, gvk.Version, gvk.Kind)
			kubeType := strings.Join(parts, "/")

			// Check if any of the regexes match, if they match, then exclude them.
			match := false
			for _, regex := range compiledRegex {
				if regex.MatchString(kubeType) {
					match = true
					break
				}
			}
			if match {
				resourceLogger(logger, r).Debugf("resource ignored")
				continue
			}

			newRes = append(newRes, r)
		}

		return newRes, nil
	}), nil
}

// NewIncludeNamespaceProcessor returns a new Resource processor that will
// remove resources that are not present in a given list of namespaces. if the
// list of namespaces is empty then all resources remain.
func NewIncludeNamespaceProcessor(includeNamespaceRegex []string, logger log.Logger) (ResourceProcessor, error) {
	logger = logger.WithValues(log.Kv{"app-svc": "process.IncludeNamespaceProcessor"})

	compiledRegex := make([]*regexp.Regexp, 0, len(includeNamespaceRegex))
	for _, r := range includeNamespaceRegex {
		if r == "" {
			continue
		}
		cr, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("invalid regex %q: %w", r, err)
		}
		compiledRegex = append(compiledRegex, cr)
	}

	return ResourceProcessorFunc(func(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {

		// If there are no regex, there is nothing to do and we return
		// all resources
		if len(compiledRegex) == 0 {
			return resources, nil
		}

		newRes := []model.Resource{}

		for _, r := range resources {
			namespace := r.K8sObject.GetNamespace()

			// Check if any of the regexes match, if they match,
			// then include them.
			match := false
			for _, regex := range compiledRegex {
				if regex.MatchString(namespace) {
					match = true
					break
				}
			}
			if !match {
				resourceLogger(logger, r).Debugf("resource ignored")
				continue
			}

			newRes = append(newRes, r)
		}

		return newRes, nil
	}), nil
}

// NewKubeLabelSelectorProcessor returns a new Resource processor that will exclude (remove)
// from the received resources, the ones that don't match with the received Kubernetes
// label selector.
func NewKubeLabelSelectorProcessor(kubeSelector string, logger log.Logger) (ResourceProcessor, error) {
	logger = logger.WithValues(log.Kv{"app-svc": "process.KubeLabelSelectorProcessor"})

	// Create label selectors.
	selectors, err := labels.ParseToRequirements(kubeSelector)
	if err != nil {
		return nil, fmt.Errorf("could not parse Kubernetes label selector %w", err)
	}

	return ResourceProcessorFunc(func(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
		newRes := []model.Resource{}

		for _, r := range resources {
			resLabels := r.K8sObject.GetLabels()
			if !matchSelector(selectors, resLabels) {
				resourceLogger(logger, r).Debugf("resource ignored")
				continue
			}
			newRes = append(newRes, r)
		}

		return newRes, nil
	}), nil
}

// NewKubeAnnotationSelectorProcessor returns a new Resource processor that will exclude (remove)
// from the received resources, the ones that don't match with the received Kubernetes
// annotation selector.
//
// Works like KubeLabelSelectorProcessor but for annotations insteado of labels.
func NewKubeAnnotationSelectorProcessor(kubeSelector string, logger log.Logger) (ResourceProcessor, error) {
	logger = logger.WithValues(log.Kv{"app-svc": "process.KubeAnnotationSelectorProcessor"})

	// Create label selectors.
	selectors, err := labels.ParseToRequirements(kubeSelector)
	if err != nil {
		return nil, fmt.Errorf("could not parse Kubernetes annotation selector %w", err)
	}

	return ResourceProcessorFunc(func(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
		newRes := []model.Resource{}

		for _, r := range resources {
			resAnnotations := r.K8sObject.GetAnnotations()
			if !matchSelector(selectors, resAnnotations) {
				resourceLogger(logger, r).Debugf("resource ignored")
				continue
			}
			newRes = append(newRes, r)
		}

		return newRes, nil
	}), nil
}

func resourceLogger(l log.Logger, r model.Resource) log.Logger {
	return l.WithValues(log.Kv{
		"resource-id":       r.ID,
		"resource-group-id": r.GroupID,
	})
}

func matchSelector(selectors []labels.Requirement, resLabels map[string]string) bool {
	labelSet := labels.Set(resLabels)
	for _, selector := range selectors {
		if !selector.Matches(labelSet) {
			return false
		}
	}

	return true
}
