package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	msqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// Config controls SQLite initialization and migrations.
type Config struct {
	Path           string
	MigrationsPath string
}

func Open(cfg Config) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, errors.New("sqlite path is required")
	}
	if cfg.MigrationsPath == "" {
		return nil, errors.New("sqlite migrations path is required")
	}

	absDBPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve db path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absDBPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_foreign_keys=on", absDBPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err := applyMigrations(db, cfg.MigrationsPath); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func applyMigrations(db *sql.DB, migrationPath string) error {
	migrationURL := migrationPath
	if !strings.HasPrefix(migrationURL, "file://") {
		absPath, err := filepath.Abs(migrationPath)
		if err != nil {
			return fmt.Errorf("resolve migrations path: %w", err)
		}
		migrationURL = "file://" + filepath.ToSlash(absPath)
	}

	driver, err := msqlite.WithInstance(db, &msqlite.Config{})
	if err != nil {
		return fmt.Errorf("create sqlite migrate driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(migrationURL, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
