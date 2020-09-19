package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

// State represents a state report of useful data and actions taken of the app execution.
type State struct {
	ID               string
	StartedAt        time.Time
	EndedAt          time.Time
	AppliedResources []Resource
	DeletedResources []Resource
}

// NewState returns a new state.
func NewState() (*State, error) {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		return nil, err
	}

	return &State{
		ID:        id.String(),
		StartedAt: t,
	}, nil
}
