package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type createGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type groupInviteRequest struct {
	GroupID   string `json:"group_id"`
	InviteeID string `json:"invitee_id"`
}

type groupJoinRequest struct {
	GroupID string `json:"group_id"`
}

type groupRespondRequest struct {
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
}

type groupItem struct {
	ID                    string  `json:"id"`
	CreatorID             string  `json:"creator_id"`
	Title                 string  `json:"title"`
	Description           string  `json:"description"`
	CreatedAt             string  `json:"created_at"`
	MemberCount           int     `json:"member_count"`
	IsMember              bool    `json:"is_member"`
	MemberRole            *string `json:"member_role,omitempty"`
	HasPendingInvite      bool    `json:"has_pending_invite"`
	HasPendingJoinRequest bool    `json:"has_pending_join_request"`
}

type groupInviteItem struct {
	ID        string   `json:"id"`
	Group     groupRef `json:"group"`
	Inviter   userCard `json:"inviter"`
	CreatedAt string   `json:"created_at"`
}

type groupJoinRequestItem struct {
	ID        string   `json:"id"`
	Group     groupRef `json:"group"`
	Requester userCard `json:"requester"`
	CreatedAt string   `json:"created_at"`
}

type groupRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (a *App) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListGroups(w, r)
	case http.MethodPost:
		a.handleCreateGroup(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *App) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" || req.Description == "" {
		writeError(w, http.StatusBadRequest, "title and description are required")
		return
	}

	groupID := uuid.NewString()
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO groups (id, creator_id, title, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, groupID, currentUser.ID, req.Title, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create group")
		return
	}

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES (?, ?, 'creator', CURRENT_TIMESTAMP)
	`, groupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create creator membership")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit group creation")
		return
	}

	group, err := a.loadGroupByID(r, currentUser.ID, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "group created but failed to load it")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]groupItem{"group": group})
}

func (a *App) handleListGroups(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	groups, err := a.loadGroupsForUser(r, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list groups")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]groupItem{"groups": groups})
}

func (a *App) handleCreateGroupInvite(w http.ResponseWriter, r *http.Request) {
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

	var req groupInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.GroupID = strings.TrimSpace(req.GroupID)
	req.InviteeID = strings.TrimSpace(req.InviteeID)
	if req.GroupID == "" || req.InviteeID == "" {
		writeError(w, http.StatusBadRequest, "group_id and invitee_id are required")
		return
	}
	if req.InviteeID == currentUser.ID {
		writeError(w, http.StatusBadRequest, "you cannot invite yourself")
		return
	}

	isMember, _, err := a.groupMembership(r, req.GroupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check group membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to invite users")
		return
	}

	if _, err := a.getUserByID(r.Context(), req.InviteeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "invitee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load invitee")
		return
	}

	inviteeFollowsInviter, err := a.isFollowing(r, req.InviteeID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check follower relationship")
		return
	}
	if !inviteeFollowsInviter {
		writeError(w, http.StatusBadRequest, "you can only invite one of your followers")
		return
	}

	alreadyMember, _, err := a.groupMembership(r, req.GroupID, req.InviteeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check invitee membership")
		return
	}
	if alreadyMember {
		writeError(w, http.StatusConflict, "invitee is already a group member")
		return
	}

	inviteID, status, err := a.upsertGroupInvite(r, req.GroupID, currentUser.ID, req.InviteeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create group invite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"invite_id": inviteID,
		"status":    status,
	})
}

func (a *App) handleIncomingGroupInvites(w http.ResponseWriter, r *http.Request) {
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
			gi.id,
			gi.group_id,
			g.title,
			CAST(gi.created_at AS TEXT),
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility
		FROM group_invites gi
		JOIN groups g ON g.id = gi.group_id
		JOIN users u ON u.id = gi.inviter_id
		WHERE gi.invitee_id = ?
			AND gi.status = 'pending'
		ORDER BY gi.created_at ASC
	`, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list incoming invites")
		return
	}
	defer rows.Close()

	invites := make([]groupInviteItem, 0)
	for rows.Next() {
		var item groupInviteItem
		var avatar sql.NullString
		var nickname sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.Group.ID,
			&item.Group.Title,
			&item.CreatedAt,
			&item.Inviter.ID,
			&item.Inviter.Email,
			&item.Inviter.FirstName,
			&item.Inviter.LastName,
			&avatar,
			&nickname,
			&item.Inviter.ProfileVisibility,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read incoming invites")
			return
		}

		item.Inviter.AvatarPath = ptrFromNull(avatar)
		item.Inviter.Nickname = ptrFromNull(nickname)
		item.Inviter.IsSelf = false
		item.Inviter.IsFollowing = false
		item.Inviter.RequestStatus = "none"
		invites = append(invites, item)
	}

	writeJSON(w, http.StatusOK, map[string][]groupInviteItem{"invites": invites})
}

func (a *App) handleRespondGroupInvite(w http.ResponseWriter, r *http.Request) {
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

	var req groupRespondRequest
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

	var groupID string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT group_id
		FROM group_invites
		WHERE id = ?
			AND invitee_id = ?
			AND status = 'pending'
		LIMIT 1
	`, req.RequestID, currentUser.ID).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pending invite not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load invite")
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		UPDATE group_invites
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newStatus, req.RequestID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update invite")
		return
	}

	if req.Action == "accept" {
		_, err = tx.ExecContext(r.Context(), `
			INSERT OR IGNORE INTO group_members (group_id, user_id, role, joined_at)
			VALUES (?, ?, 'member', CURRENT_TIMESTAMP)
		`, groupID, currentUser.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create membership")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit invite response")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"request_id": req.RequestID,
		"status":     newStatus,
	})
}

func (a *App) handleCreateGroupJoinRequest(w http.ResponseWriter, r *http.Request) {
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

	var req groupJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.GroupID = strings.TrimSpace(req.GroupID)
	if req.GroupID == "" {
		writeError(w, http.StatusBadRequest, "group_id is required")
		return
	}

	if _, err := a.groupExists(r, req.GroupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load group")
		return
	}

	alreadyMember, _, err := a.groupMembership(r, req.GroupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if alreadyMember {
		writeError(w, http.StatusConflict, "you are already a member of this group")
		return
	}

	requestID, status, err := a.upsertGroupJoinRequest(r, req.GroupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create join request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"request_id": requestID,
		"status":     status,
	})
}

func (a *App) handleIncomingGroupJoinRequests(w http.ResponseWriter, r *http.Request) {
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
			gjr.id,
			gjr.group_id,
			g.title,
			CAST(gjr.created_at AS TEXT),
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname,
			u.profile_visibility
		FROM group_join_requests gjr
		JOIN groups g ON g.id = gjr.group_id
		JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = ? AND gm.role = 'creator'
		JOIN users u ON u.id = gjr.requester_id
		WHERE gjr.status = 'pending'
		ORDER BY gjr.created_at ASC
	`, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list join requests")
		return
	}
	defer rows.Close()

	requests := make([]groupJoinRequestItem, 0)
	for rows.Next() {
		var item groupJoinRequestItem
		var avatar sql.NullString
		var nickname sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.Group.ID,
			&item.Group.Title,
			&item.CreatedAt,
			&item.Requester.ID,
			&item.Requester.Email,
			&item.Requester.FirstName,
			&item.Requester.LastName,
			&avatar,
			&nickname,
			&item.Requester.ProfileVisibility,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read join requests")
			return
		}

		item.Requester.AvatarPath = ptrFromNull(avatar)
		item.Requester.Nickname = ptrFromNull(nickname)
		item.Requester.IsSelf = false
		item.Requester.IsFollowing = false
		item.Requester.RequestStatus = "none"
		requests = append(requests, item)
	}

	writeJSON(w, http.StatusOK, map[string][]groupJoinRequestItem{"requests": requests})
}

func (a *App) handleRespondGroupJoinRequest(w http.ResponseWriter, r *http.Request) {
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

	var req groupRespondRequest
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

	var groupID string
	var requesterID string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT gjr.group_id, gjr.requester_id
		FROM group_join_requests gjr
		JOIN group_members gm ON gm.group_id = gjr.group_id AND gm.user_id = ? AND gm.role = 'creator'
		WHERE gjr.id = ?
			AND gjr.status = 'pending'
		LIMIT 1
	`, currentUser.ID, req.RequestID).Scan(&groupID, &requesterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pending join request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load join request")
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		UPDATE group_join_requests
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newStatus, req.RequestID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update join request")
		return
	}

	if req.Action == "accept" {
		_, err = tx.ExecContext(r.Context(), `
			INSERT OR IGNORE INTO group_members (group_id, user_id, role, joined_at)
			VALUES (?, ?, 'member', CURRENT_TIMESTAMP)
		`, groupID, requesterID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to add group member")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit join response")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"request_id": req.RequestID,
		"status":     newStatus,
	})
}

func (a *App) loadGroupsForUser(r *http.Request, userID string) ([]groupItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			g.id,
			g.creator_id,
			g.title,
			g.description,
			CAST(g.created_at AS TEXT),
			(
				SELECT COUNT(*)
				FROM group_members gm_count
				WHERE gm_count.group_id = g.id
			) AS member_count,
			COALESCE(gm.role, '') AS member_role,
			EXISTS(
				SELECT 1
				FROM group_invites gi
				WHERE gi.group_id = g.id
					AND gi.invitee_id = ?
					AND gi.status = 'pending'
			) AS has_pending_invite,
			EXISTS(
				SELECT 1
				FROM group_join_requests gjr
				WHERE gjr.group_id = g.id
					AND gjr.requester_id = ?
					AND gjr.status = 'pending'
			) AS has_pending_join_request
		FROM groups g
		LEFT JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = ?
		ORDER BY g.created_at DESC
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]groupItem, 0)
	for rows.Next() {
		var item groupItem
		var memberRole string
		if err := rows.Scan(
			&item.ID,
			&item.CreatorID,
			&item.Title,
			&item.Description,
			&item.CreatedAt,
			&item.MemberCount,
			&memberRole,
			&item.HasPendingInvite,
			&item.HasPendingJoinRequest,
		); err != nil {
			return nil, err
		}

		if memberRole != "" {
			item.IsMember = true
			item.MemberRole = &memberRole
		}

		groups = append(groups, item)
	}

	return groups, rows.Err()
}

func (a *App) loadGroupByID(r *http.Request, userID, groupID string) (groupItem, error) {
	groups, err := a.loadGroupsForUser(r, userID)
	if err != nil {
		return groupItem{}, err
	}

	for _, group := range groups {
		if group.ID == groupID {
			return group, nil
		}
	}

	return groupItem{}, sql.ErrNoRows
}

func (a *App) groupMembership(r *http.Request, groupID, userID string) (bool, string, error) {
	var role string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT role
		FROM group_members
		WHERE group_id = ?
			AND user_id = ?
		LIMIT 1
	`, groupID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}

	return true, role, nil
}

func (a *App) groupExists(r *http.Request, groupID string) (bool, error) {
	var exists bool
	err := a.db.QueryRowContext(r.Context(), `
		SELECT EXISTS(
			SELECT 1
			FROM groups
			WHERE id = ?
		)
	`, groupID).Scan(&exists)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, sql.ErrNoRows
	}
	return true, nil
}

func (a *App) upsertGroupInvite(r *http.Request, groupID, inviterID, inviteeID string) (string, string, error) {
	var existingID string
	var existingStatus string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT id, status
		FROM group_invites
		WHERE group_id = ?
			AND invitee_id = ?
		LIMIT 1
	`, groupID, inviteeID).Scan(&existingID, &existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID := uuid.NewString()
		_, err := a.db.ExecContext(r.Context(), `
			INSERT INTO group_invites (id, group_id, inviter_id, invitee_id, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, newID, groupID, inviterID, inviteeID)
		if err != nil {
			return "", "", err
		}
		return newID, "pending", nil
	}

	if existingStatus == "pending" {
		return existingID, "pending", nil
	}

	_, err = a.db.ExecContext(r.Context(), `
		UPDATE group_invites
		SET inviter_id = ?, status = 'pending', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, inviterID, existingID)
	if err != nil {
		return "", "", err
	}

	return existingID, "pending", nil
}

func (a *App) upsertGroupJoinRequest(r *http.Request, groupID, requesterID string) (string, string, error) {
	var existingID string
	var existingStatus string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT id, status
		FROM group_join_requests
		WHERE group_id = ?
			AND requester_id = ?
		LIMIT 1
	`, groupID, requesterID).Scan(&existingID, &existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID := uuid.NewString()
		_, err := a.db.ExecContext(r.Context(), `
			INSERT INTO group_join_requests (id, group_id, requester_id, status, created_at, updated_at)
			VALUES (?, ?, ?, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, newID, groupID, requesterID)
		if err != nil {
			return "", "", err
		}
		return newID, "pending", nil
	}

	if existingStatus == "pending" {
		return existingID, "pending", nil
	}

	_, err = a.db.ExecContext(r.Context(), `
		UPDATE group_join_requests
		SET status = 'pending', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, existingID)
	if err != nil {
		return "", "", err
	}

	return existingID, "pending", nil
}
