package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// App wires HTTP routes with shared dependencies.
type App struct {
	db              *sql.DB
	uploadDir       string
	sessionDuration time.Duration
}

// Config controls runtime app behavior.
type Config struct {
	UploadDir       string
	SessionDuration time.Duration
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Time     string `json:"time"`
}

func New(db *sql.DB, cfg Config) (*App, error) {
	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./data/uploads"
	}

	sessionDuration := cfg.SessionDuration
	if sessionDuration <= 0 {
		sessionDuration = 7 * 24 * time.Hour
	}

	if err := os.MkdirAll(filepath.Join(uploadDir, "avatars"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(uploadDir, "posts"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(uploadDir, "comments"), 0o755); err != nil {
		return nil, err
	}

	return &App{
		db:              db,
		uploadDir:       uploadDir,
		sessionDuration: sessionDuration,
	}, nil
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/auth/register", a.handleRegister)
	mux.HandleFunc("/api/auth/login", a.handleLogin)
	mux.HandleFunc("/api/auth/logout", a.handleLogout)
	mux.HandleFunc("/api/auth/me", a.handleMe)
	mux.HandleFunc("/api/users", a.handleListUsers)
	mux.HandleFunc("/api/follows", a.handleMyFollows)
	mux.HandleFunc("/api/follows/request", a.handleCreateFollowRequest)
	mux.HandleFunc("/api/follows/unfollow", a.handleUnfollow)
	mux.HandleFunc("/api/follows/requests/incoming", a.handleIncomingFollowRequests)
	mux.HandleFunc("/api/follows/requests/respond", a.handleRespondFollowRequest)
	mux.HandleFunc("/api/profile/me", a.handleMyProfile)
	mux.HandleFunc("/api/profile/view", a.handleViewProfile)
	mux.HandleFunc("/api/profile/me/visibility", a.handlePatchProfileVisibility)
	mux.HandleFunc("/api/posts", a.handleCreatePost)
	mux.HandleFunc("/api/posts/feed", a.handleFeedPosts)
	mux.HandleFunc("/api/posts/comments", a.handleCreateComment)
	mux.HandleFunc("/api/groups", a.handleGroups)
	mux.HandleFunc("/api/groups/invites", a.handleCreateGroupInvite)
	mux.HandleFunc("/api/groups/invites/incoming", a.handleIncomingGroupInvites)
	mux.HandleFunc("/api/groups/invites/respond", a.handleRespondGroupInvite)
	mux.HandleFunc("/api/groups/requests/join", a.handleCreateGroupJoinRequest)
	mux.HandleFunc("/api/groups/requests/incoming", a.handleIncomingGroupJoinRequests)
	mux.HandleFunc("/api/groups/requests/respond", a.handleRespondGroupJoinRequest)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(a.uploadDir))))
	return withCORS(mux)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := healthResponse{
		Status:   "ok",
		Database: "ok",
		Time:     time.Now().UTC().Format(time.RFC3339),
	}

	if err := a.db.PingContext(r.Context()); err != nil {
		response.Status = "degraded"
		response.Database = "error"
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func isUnauthorizedError(err error) bool {
	return errors.Is(err, errUnauthorized)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
