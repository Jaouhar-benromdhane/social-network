package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── Rate Limiter ────────────────────────────────────────────────────────────

type rateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		hits:   make(map[string][]time.Time),
		max:    max,
		window: window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Keep only hits within the window
	filtered := rl.hits[key][:0]
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	rl.hits[key] = filtered

	if len(rl.hits[key]) >= rl.max {
		return false
	}
	rl.hits[key] = append(rl.hits[key], now)
	return true
}

// ─── Sanitize ────────────────────────────────────────────────────────────────

// sanitizeInput strips HTML tags from user-supplied text to prevent stored XSS.
func sanitizeInput(s string) string {
	// html.EscapeString converts < > & " ' so scripts can never be executed.
	return html.EscapeString(strings.TrimSpace(s))
}

// App wires HTTP routes with shared dependencies.
type App struct {
	db              *sql.DB
	uploadDir       string
	sessionDuration time.Duration
	wsHub           *wsHub
	loginLimiter    *rateLimiter
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
		wsHub:           newWSHub(),
		// 10 login attempts per minute per IP
		loginLimiter: newRateLimiter(10, time.Minute),
	}, nil
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/auth/register", a.handleRegister)
	mux.HandleFunc("/api/auth/login", a.rateLimitedLogin)
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
	mux.HandleFunc("/api/groups/posts", a.handleGroupPosts)
	mux.HandleFunc("/api/groups/posts/comments", a.handleCreateGroupComment)
	mux.HandleFunc("/api/groups/events", a.handleGroupEvents)
	mux.HandleFunc("/api/groups/events/vote", a.handleVoteGroupEvent)
	mux.HandleFunc("/api/ws", a.handleWebSocket)
	mux.HandleFunc("/api/chat/private/messages", a.handlePrivateMessages)
	mux.HandleFunc("/api/chat/groups/messages", a.handleGroupMessages)
	mux.HandleFunc("/api/notifications", a.handleNotifications)
	mux.HandleFunc("/api/notifications/read", a.handleMarkNotificationsRead)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(a.uploadDir))))
	return withSecurityHeaders(withCORS(mux))
}

// rateLimitedLogin wraps handleLogin with IP-based rate limiting.
func (a *App) rateLimitedLogin(w http.ResponseWriter, r *http.Request) {
	// Extract IP (strip port)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	if !a.loginLimiter.allow(ip) {
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, "too many login attempts, please try again in a minute")
		return
	}
	a.handleLogin(w, r)
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

// allowedOrigins lists the origins that are permitted to make credentialed
// cross-origin requests. Adjust for staging/production as needed.
var allowedOrigins = map[string]bool{
	"http://localhost:3000":  true,
	"http://127.0.0.1:3000": true,
	"http://localhost:5173":  true, // Vite dev server
	"http://127.0.0.1:5173": true,
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && allowedOrigins[origin] {
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

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
