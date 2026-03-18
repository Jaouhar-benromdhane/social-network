package config

import (
	"os"
	"strconv"
)

// Config holds application runtime settings loaded from environment variables.
type Config struct {
	Port           string
	DBPath         string
	MigrationsPath string
	UploadDir      string
	SessionHours   int
}

func Load() Config {
	return Config{
		Port:           envOrDefault("APP_PORT", "8080"),
		DBPath:         envOrDefault("DB_PATH", "./data/social_network.db"),
		MigrationsPath: envOrDefault("MIGRATIONS_PATH", "./pkg/db/migrations/sqlite"),
		UploadDir:      envOrDefault("UPLOAD_DIR", "./data/uploads"),
		SessionHours:   envOrDefaultInt("SESSION_HOURS", 168),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envOrDefaultInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
