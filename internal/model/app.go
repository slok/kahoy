package model

import (
	"context"
	"time"
)

// AppConfig is the configuration of the app.
type AppConfig struct {
	// Group configuration by ID
	Groups map[string]GroupConfig
}

// GroupConfig is the group configuration.
type GroupConfig struct {
	Priority   *int
	WaitConfig *GroupWaitConfig
}

// GroupWaitConfig has a group wait options.
type GroupWaitConfig struct {
	Duration time.Duration
}

// Validate will validate the app configuration.
func (c *AppConfig) Validate(ctx context.Context) error {
	if c.Groups == nil {
		c.Groups = map[string]GroupConfig{}
	}
	return nil
}
