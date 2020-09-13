package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/kahoy/internal/model"
)

func TestNewReport(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Check a new report has ID and time set when is created.
	report1, err := model.NewReport()
	require.NoError(err)
	assert.NotEmpty(report1.ID)
	assert.NotEmpty(report1.StartedAt)
	assert.Empty(report1.EndedAt)

	report2, err := model.NewReport()
	require.NoError(err)
	assert.NotEmpty(report2.ID)
	assert.NotEmpty(report2.StartedAt)
	assert.Empty(report2.EndedAt)

	// Check report1 and report2 are different.
	assert.NotEqual(report1.ID, report2.ID)
	assert.NotEqual(report1.StartedAt, report2.StartedAt)
}
