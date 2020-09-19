package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
)

func TestNewState(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Check a new state has ID and time set when is created.
	state1, err := model.NewState()
	require.NoError(err)
	assert.NotEmpty(state1.ID)
	assert.NotEmpty(state1.StartedAt)
	assert.Empty(state1.EndedAt)

	state2, err := model.NewState()
	require.NoError(err)
	assert.NotEmpty(state2.ID)
	assert.NotEmpty(state2.StartedAt)
	assert.Empty(state2.EndedAt)

	// Check state1 and state2 are different.
	assert.NotEqual(state1.ID, state2.ID)
	assert.NotEqual(state1.StartedAt, state2.StartedAt)
}
