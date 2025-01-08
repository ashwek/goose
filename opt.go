package goose

type MigrationConfig struct {
	Scope string
}

type MigrationOption func(cfg *MigrationConfig)

func WithScope(scope string) MigrationOption {
	return func(cfg *MigrationConfig) {
		cfg.Scope = scope
	}
}
