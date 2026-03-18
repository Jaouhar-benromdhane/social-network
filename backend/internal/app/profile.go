package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type profileStats struct {
	Followers int `json:"followers"`
	Following int `json:"following"`
}

type myProfileResponse struct {
	User      User         `json:"user"`
	Stats     profileStats `json:"stats"`
	Followers []userCard   `json:"followers"`
	Following []userCard   `json:"following"`
	Posts     []postItem   `json:"posts"`
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

	followers, err := a.loadFollowers(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load followers")
		return
	}

	following, err := a.loadFollowing(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load following")
		return
	}

	posts, err := a.loadVisiblePosts(r, user.ID, &user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load posts")
		return
	}

	writeJSON(w, http.StatusOK, myProfileResponse{
		User:      user,
		Stats:     stats,
		Followers: followers,
		Following: following,
		Posts:     posts,
	})
}

func (a *App) handleViewProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	viewer, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	targetID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if targetID == "" {
		writeError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	targetUser, err := a.getUserByID(r.Context(), targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load target user")
		return
	}

	canView, err := a.canViewUserProfile(r, viewer.ID, targetUser.ID, targetUser.ProfileVisibility)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check profile permissions")
		return
	}
	if !canView {
		writeError(w, http.StatusForbidden, "private profile: follow this user to see details")
		return
	}

	stats, err := a.loadProfileStats(r, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load profile stats")
		return
	}

	followers, err := a.loadFollowers(r, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load followers")
		return
	}

	following, err := a.loadFollowing(r, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load following")
		return
	}

	posts, err := a.loadVisiblePosts(r, viewer.ID, &targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load posts")
		return
	}

	writeJSON(w, http.StatusOK, myProfileResponse{
		User:      targetUser,
		Stats:     stats,
		Followers: followers,
		Following: following,
		Posts:     posts,
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
