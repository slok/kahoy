package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

// Report represents a report of useful data and actions taken of the app execution.
type Report struct {
	ID               string
	StartedAt        time.Time
	EndedAt          time.Time
	AppliedResources []Resource
	DeletedResources []Resource
}

// NewReport returns a new report.
func NewReport() (*Report, error) {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		return nil, err
	}

	return &Report{
		ID:        id.String(),
		StartedAt: t,
	}, nil
}
