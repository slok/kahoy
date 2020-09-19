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

type stateRepository struct {
	out io.Writer
}

// NewStateRepository returns a new repository that knows how to write JSON
// states on the received output.
func NewStateRepository(out io.Writer) storage.StateRepository {
	return stateRepository{out: out}
}

func (s stateRepository) StoreState(ctx context.Context, state model.State) error {
	data, err := s.mapStateToJSON(state)
	if err != nil {
		return fmt.Errorf("could not map state to JSON: %w", err)
	}

	_, err = s.out.Write(data)
	if err != nil {
		return fmt.Errorf("could not write JSON state: %w", err)
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

func (s stateRepository) mapStateToJSON(state model.State) ([]byte, error) {
	// Map resources.
	applied := make([]jsonResource, 0, len(state.AppliedResources))
	for _, res := range state.AppliedResources {
		applied = append(applied, s.mapResourceToJSON(res))
	}
	deleted := make([]jsonResource, 0, len(state.DeletedResources))
	for _, res := range state.DeletedResources {
		deleted = append(deleted, s.mapResourceToJSON(res))
	}

	jr := jsonReport{
		Version:          "v1",
		ID:               state.ID,
		StartedAt:        state.StartedAt.Format(time.RFC3339),
		EndedAt:          state.EndedAt.Format(time.RFC3339),
		AppliedResources: applied,
		DeletedResources: deleted,
	}

	data, err := json.Marshal(jr)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s stateRepository) mapResourceToJSON(res model.Resource) jsonResource {
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
