package model

// AppConfig is the configuration of the app.
type AppConfig struct {
	// Group configuration by ID
	Groups map[string]GroupConfig
}

// GroupConfig is the group confdiguration.
type GroupConfig struct {
	Priority *int
}
