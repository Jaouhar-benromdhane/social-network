package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type followTargetRequest struct {
	TargetUserID string `json:"target_user_id"`
}

type followRespondRequest struct {
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
}

type userCard struct {
	ID                string  `json:"id"`
	Email             string  `json:"email"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	AvatarPath        *string `json:"avatar_path,omitempty"`
	Nickname          *string `json:"nickname,omitempty"`
	ProfileVisibility string  `json:"profile_visibility"`
	IsSelf            bool    `json:"is_self"`
	IsFollowing       bool    `json:"is_following"`
	RequestStatus     string  `json:"request_status"`
}

type followRequestItem struct {
	ID        string   `json:"id"`
	Requester userCard `json:"requester"`
	CreatedAt string   `json:"created_at"`
}

type followsResponse struct {
	Followers []userCard `json:"followers"`
	Following []userCard `json:"following"`
}

func (a *App) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility,
			EXISTS(
				SELECT 1
				FROM follows f
				WHERE f.follower_id = ?
					AND f.following_id = u.id
			) AS is_following,
			COALESCE((
				SELECT fr.status
				FROM follow_requests fr
				WHERE fr.requester_id = ?
					AND fr.target_id = u.id
				ORDER BY fr.created_at DESC
				LIMIT 1
			), '') AS request_status
		FROM users u
		ORDER BY u.created_at DESC
	`, currentUser.ID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	users := make([]userCard, 0)
	for rows.Next() {
		var card userCard
		var avatar sql.NullString
		var nickname sql.NullString
		var requestStatus string
		if err := rows.Scan(
			&card.ID,
			&card.Email,
			&card.FirstName,
			&card.LastName,
			&avatar,
			&nickname,
			&card.ProfileVisibility,
			&card.IsFollowing,
			&requestStatus,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read users")
			return
		}

		card.AvatarPath = ptrFromNull(avatar)
		card.Nickname = ptrFromNull(nickname)
		card.IsSelf = card.ID == currentUser.ID
		if requestStatus == "" {
			requestStatus = "none"
		}
		card.RequestStatus = requestStatus
		users = append(users, card)
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (a *App) handleCreateFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req followTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.TargetUserID = strings.TrimSpace(req.TargetUserID)
	if req.TargetUserID == "" {
		writeError(w, http.StatusBadRequest, "target_user_id is required")
		return
	}
	if req.TargetUserID == currentUser.ID {
		writeError(w, http.StatusBadRequest, "you cannot follow yourself")
		return
	}

	targetUser, err := a.getUserByID(r.Context(), req.TargetUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "target user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load target user")
		return
	}

	if isFollowing, err := a.isFollowing(r, currentUser.ID, targetUser.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check follow status")
		return
	} else if isFollowing {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":        "already_following",
			"target_user":   targetUser,
			"auto_followed": false,
		})
		return
	}

	if targetUser.ProfileVisibility == "public" {
		if err := a.createFollow(r, currentUser.ID, targetUser.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to follow user")
			return
		}
		_, _ = a.db.ExecContext(r.Context(), `
			UPDATE follow_requests
			SET status = 'accepted', updated_at = CURRENT_TIMESTAMP
			WHERE requester_id = ?
				AND target_id = ?
				AND status = 'pending'
		`, currentUser.ID, targetUser.ID)

		requesterName := strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName)
		if requesterName == "" {
			requesterName = currentUser.Email
		}
		_ = a.pushNotification(r.Context(), targetUser.ID, "follow_request", map[string]any{
			"requester_id":       currentUser.ID,
			"requester_name":     requesterName,
			"requester_nickname": currentUser.Nickname,
			"auto_followed":      true,
		})

		writeJSON(w, http.StatusOK, map[string]any{
			"status":        "following",
			"auto_followed": true,
			"target_user":   targetUser,
		})
		return
	}

	requestID, status, err := a.upsertFollowRequest(r, currentUser.ID, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create follow request")
		return
	}

	_ = a.pushNotification(r.Context(), targetUser.ID, "follow_request", map[string]any{
		"request_id":         requestID,
		"requester_id":       currentUser.ID,
		"requester_name":     strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName),
		"requester_nickname": currentUser.Nickname,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status":           status,
		"request_id":       requestID,
		"target_user":      targetUser,
		"auto_followed":    false,
		"target_is_public": false,
	})
}

func (a *App) handleIncomingFollowRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			fr.id,
			CAST(fr.created_at AS TEXT),
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility
		FROM follow_requests fr
		JOIN users u ON u.id = fr.requester_id
		WHERE fr.target_id = ?
			AND fr.status = 'pending'
		ORDER BY fr.created_at ASC
	`, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list incoming requests")
		return
	}
	defer rows.Close()

	requests := make([]followRequestItem, 0)
	for rows.Next() {
		var item followRequestItem
		var avatar sql.NullString
		var nickname sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.CreatedAt,
			&item.Requester.ID,
			&item.Requester.Email,
			&item.Requester.FirstName,
			&item.Requester.LastName,
			&avatar,
			&nickname,
			&item.Requester.ProfileVisibility,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read incoming requests")
			return
		}

		item.Requester.AvatarPath = ptrFromNull(avatar)
		item.Requester.Nickname = ptrFromNull(nickname)
		item.Requester.IsSelf = false
		item.Requester.IsFollowing = false
		item.Requester.RequestStatus = "pending"
		requests = append(requests, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{"requests": requests})
}

func (a *App) handleRespondFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req followRespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.RequestID = strings.TrimSpace(req.RequestID)
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	if req.RequestID == "" {
		writeError(w, http.StatusBadRequest, "request_id is required")
		return
	}
	if req.Action != "accept" && req.Action != "decline" {
		writeError(w, http.StatusBadRequest, "action must be accept or decline")
		return
	}

	newStatus := "declined"
	if req.Action == "accept" {
		newStatus = "accepted"
	}

	var requesterID string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT requester_id
		FROM follow_requests
		WHERE id = ?
			AND target_id = ?
			AND status = 'pending'
		LIMIT 1
	`, req.RequestID, currentUser.ID).Scan(&requesterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pending request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load follow request")
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		UPDATE follow_requests
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newStatus, req.RequestID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update follow request")
		return
	}

	if req.Action == "accept" {
		_, err = tx.ExecContext(r.Context(), `
			INSERT OR IGNORE INTO follows (follower_id, following_id, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, requesterID, currentUser.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create follow link")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit follow response")
		return
	}

	responderName := strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName)
	if responderName == "" {
		responderName = currentUser.Email
	}

	if req.Action == "accept" {
		_ = a.pushNotification(r.Context(), requesterID, "follow_request_accepted", map[string]any{
			"target_id":   currentUser.ID,
			"target_name": responderName,
		})
	} else {
		_ = a.pushNotification(r.Context(), requesterID, "follow_request_declined", map[string]any{
			"target_id":   currentUser.ID,
			"target_name": responderName,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"request_id": req.RequestID,
		"status":     newStatus,
	})
}

func (a *App) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req followTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.TargetUserID = strings.TrimSpace(req.TargetUserID)
	if req.TargetUserID == "" {
		writeError(w, http.StatusBadRequest, "target_user_id is required")
		return
	}

	result, err := a.db.ExecContext(r.Context(), `
		DELETE FROM follows
		WHERE follower_id = ?
			AND following_id = ?
	`, currentUser.ID, req.TargetUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unfollow user")
		return
	}

	affected, _ := result.RowsAffected()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "unfollowed",
		"target_user_id": req.TargetUserID,
		"deleted":        affected > 0,
	})
}

func (a *App) handleMyFollows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	followers, err := a.loadFollowers(r, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load followers")
		return
	}

	following, err := a.loadFollowing(r, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load following")
		return
	}

	writeJSON(w, http.StatusOK, followsResponse{
		Followers: followers,
		Following: following,
	})
}

func (a *App) loadFollowers(r *http.Request, userID string) ([]userCard, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility
		FROM follows f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = ?
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followers := make([]userCard, 0)
	for rows.Next() {
		var card userCard
		var avatar sql.NullString
		var nickname sql.NullString
		if err := rows.Scan(
			&card.ID,
			&card.Email,
			&card.FirstName,
			&card.LastName,
			&avatar,
			&nickname,
			&card.ProfileVisibility,
		); err != nil {
			return nil, err
		}
		card.AvatarPath = ptrFromNull(avatar)
		card.Nickname = ptrFromNull(nickname)
		card.IsSelf = false
		card.IsFollowing = true
		card.RequestStatus = "none"
		followers = append(followers, card)
	}

	return followers, nil
}

func (a *App) loadFollowing(r *http.Request, userID string) ([]userCard, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility
		FROM follows f
		JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	following := make([]userCard, 0)
	for rows.Next() {
		var card userCard
		var avatar sql.NullString
		var nickname sql.NullString
		if err := rows.Scan(
			&card.ID,
			&card.Email,
			&card.FirstName,
			&card.LastName,
			&avatar,
			&nickname,
			&card.ProfileVisibility,
		); err != nil {
			return nil, err
		}
		card.AvatarPath = ptrFromNull(avatar)
		card.Nickname = ptrFromNull(nickname)
		card.IsSelf = false
		card.IsFollowing = true
		card.RequestStatus = "none"
		following = append(following, card)
	}

	return following, nil
}

func (a *App) isFollowing(r *http.Request, followerID, followingID string) (bool, error) {
	var exists bool
	err := a.db.QueryRowContext(r.Context(), `
		SELECT EXISTS(
			SELECT 1
			FROM follows
			WHERE follower_id = ?
				AND following_id = ?
		)
	`, followerID, followingID).Scan(&exists)
	return exists, err
}

func (a *App) createFollow(r *http.Request, followerID, followingID string) error {
	_, err := a.db.ExecContext(r.Context(), `
		INSERT OR IGNORE INTO follows (follower_id, following_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, followerID, followingID)
	return err
}

func (a *App) upsertFollowRequest(r *http.Request, requesterID, targetID string) (string, string, error) {
	var existingID string
	var existingStatus string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT id, status
		FROM follow_requests
		WHERE requester_id = ?
			AND target_id = ?
		LIMIT 1
	`, requesterID, targetID).Scan(&existingID, &existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID := uuid.NewString()
		_, err := a.db.ExecContext(r.Context(), `
			INSERT INTO follow_requests (id, requester_id, target_id, status, created_at, updated_at)
			VALUES (?, ?, ?, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, newID, requesterID, targetID)
		if err != nil {
			return "", "", err
		}
		return newID, "pending", nil
	}

	if existingStatus == "pending" {
		return existingID, "pending", nil
	}

	_, err = a.db.ExecContext(r.Context(), `
		UPDATE follow_requests
		SET status = 'pending', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, existingID)
	if err != nil {
		return "", "", err
	}

	return existingID, "pending", nil
}

func (a *App) canViewUserProfile(r *http.Request, viewerID, targetID, targetVisibility string) (bool, error) {
	if viewerID == targetID {
		return true, nil
	}
	if targetVisibility == "public" {
		return true, nil
	}
	return a.isFollowing(r, viewerID, targetID)
}

func (a *App) requestNowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
