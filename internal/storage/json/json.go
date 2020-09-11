package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/storage"
)

type reportRepository struct {
	out io.Writer
}

// NewReportRepository returns a new repository that knows how to write JSON
// reports on the received output.
func NewReportRepository(out io.Writer) storage.ReportRepository {
	return reportRepository{out: out}
}

func (r reportRepository) StoreReport(ctx context.Context, report model.Report) error {
	data, err := r.mapReportToJSON(report)
	if err != nil {
		return fmt.Errorf("could not map report to JSON: %w", err)
	}

	_, err = r.out.Write(data)
	if err != nil {
		return fmt.Errorf("could not write JSON report: %w", err)
	}

	return nil
}

type jsonReport struct {
	Version string `json:"version"`
	ID      string `json:"id"`
	// Representation in RFC3339.
	StartedAt string `json:"started_at"`
	// Representation in RFC3339.
	EndedAt          string         `json:"ended_at"`
	AppliedResources []jsonResource `json:"applied_resources"`
	DeletedResources []jsonResource `json:"deleted_resources"`
}

type jsonResource struct {
	ID         string `json:"id"`
	Group      string `json:"group"`
	GVK        string `json:"gvk"`
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
}

func (r reportRepository) mapReportToJSON(report model.Report) ([]byte, error) {
	// Map resources.
	applied := make([]jsonResource, 0, len(report.AppliedResources))
	for _, res := range report.AppliedResources {
		applied = append(applied, r.mapResourceToJSON(res))
	}
	deleted := make([]jsonResource, 0, len(report.DeletedResources))
	for _, res := range report.DeletedResources {
		deleted = append(deleted, r.mapResourceToJSON(res))
	}

	jr := jsonReport{
		Version:          "v1",
		ID:               report.ID,
		StartedAt:        report.StartedAt.Format(time.RFC3339),
		EndedAt:          report.EndedAt.Format(time.RFC3339),
		AppliedResources: applied,
		DeletedResources: deleted,
	}

	data, err := json.Marshal(jr)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r reportRepository) mapResourceToJSON(res model.Resource) jsonResource {
	gvk := res.K8sObject.GetObjectKind().GroupVersionKind()

	jr := jsonResource{
		ID:         res.ID,
		Group:      res.GroupID,
		GVK:        strings.Join([]string{gvk.Group, gvk.Version, gvk.Kind}, "/"),
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Namespace:  res.K8sObject.GetNamespace(),
		Name:       res.K8sObject.GetName(),
	}

	return jr
}
