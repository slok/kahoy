package process

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

func resourceLogger(l log.Logger, r model.Resource) log.Logger {
	return l.WithValues(log.Kv{
		"resource-id":       r.ID,
		"resource-group-id": r.GroupID,
	})
}
