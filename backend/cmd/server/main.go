package main

import (
	"log"
	"net/http"
	"time"

	"social-network/backend/internal/app"
	"social-network/backend/internal/config"
	dbsqlite "social-network/backend/pkg/db/sqlite"
)

func main() {
	cfg := config.Load()

	db, err := dbsqlite.Open(dbsqlite.Config{
		Path:           cfg.DBPath,
		MigrationsPath: cfg.MigrationsPath,
	})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	application, err := app.New(db, app.Config{
		UploadDir:       cfg.UploadDir,
		SessionDuration: time.Duration(cfg.SessionHours) * time.Hour,
	})
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           application.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("backend listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
