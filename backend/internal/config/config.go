package config

import "os"

// Config holds application runtime settings loaded from environment variables.
type Config struct {
	Port           string
	DBPath         string
	MigrationsPath string
}

func Load() Config {
	return Config{
		Port:           envOrDefault("APP_PORT", "8080"),
		DBPath:         envOrDefault("DB_PATH", "./data/social_network.db"),
		MigrationsPath: envOrDefault("MIGRATIONS_PATH", "./pkg/db/migrations/sqlite"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
