package configuration

import (
	"context"

	"github.com/slok/kahoy/internal/model"
)

// Loader knows how to load app configuration in different formats and sources.
type Loader interface {
	Load(ctx context.Context) (*model.AppConfig, error)
}

// LoaderFunc is a helper to create configuration loaders without declaring a new type.
type LoaderFunc func(ctx context.Context) (*model.AppConfig, error)

var _ Loader = LoaderFunc(nil)

// Load satisifies configuration.Loader interface.
func (l LoaderFunc) Load(ctx context.Context) (*model.AppConfig, error) { return l(ctx) }
