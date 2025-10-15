package database

import (
	"fmt"

	"github.com/frahmantamala/jadiles/internal"
)

func BuildPostgresSource(cfg internal.DatabaseConfig) string {
	if cfg.URL != "" {
		return cfg.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}
