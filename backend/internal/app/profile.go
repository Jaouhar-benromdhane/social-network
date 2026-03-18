package app

import (
	"encoding/json"
	"net/http"
	"strings"
)

type profileStats struct {
	Followers int `json:"followers"`
	Following int `json:"following"`
}

type myProfileResponse struct {
	User  User         `json:"user"`
	Stats profileStats `json:"stats"`
	Posts []any        `json:"posts"`
}

type visibilityRequest struct {
	Visibility string `json:"visibility"`
}

func (a *App) handleMyProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	user, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	stats, err := a.loadProfileStats(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load profile stats")
		return
	}

	writeJSON(w, http.StatusOK, myProfileResponse{
		User:  user,
		Stats: stats,
		Posts: []any{},
	})
}

func (a *App) handlePatchProfileVisibility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		methodNotAllowed(w)
		return
	}

	user, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	var req visibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Visibility = strings.TrimSpace(req.Visibility)
	if req.Visibility != "public" && req.Visibility != "private" {
		writeError(w, http.StatusBadRequest, "visibility must be public or private")
		return
	}

	_, err = a.db.ExecContext(r.Context(), `
		UPDATE users
		SET profile_visibility = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Visibility, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update visibility")
		return
	}

	updated, err := a.getUserByID(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load updated profile")
		return
	}

	writeJSON(w, http.StatusOK, map[string]User{"user": updated})
}

func (a *App) loadProfileStats(r *http.Request, userID string) (profileStats, error) {
	var followers int
	var following int

	if err := a.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*)
		FROM follows
		WHERE following_id = ?
	`, userID).Scan(&followers); err != nil {
		return profileStats{}, err
	}

	if err := a.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = ?
	`, userID).Scan(&following); err != nil {
		return profileStats{}, err
	}

	return profileStats{Followers: followers, Following: following}, nil
}
