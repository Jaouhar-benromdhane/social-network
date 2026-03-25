package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type privateMessageRequest struct {
	RecipientID string `json:"recipient_id"`
	Content     string `json:"content"`
}

type groupMessageRequest struct {
	GroupID string `json:"group_id"`
	Content string `json:"content"`
}

type chatMessageItem struct {
	ID          string     `json:"id"`
	SenderID    string     `json:"sender_id"`
	RecipientID *string    `json:"recipient_id,omitempty"`
	GroupID     *string    `json:"group_id,omitempty"`
	Content     string     `json:"content"`
	CreatedAt   string     `json:"created_at"`
	Sender      postAuthor `json:"sender"`
}

func (a *App) handlePrivateMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListPrivateMessages(w, r)
	case http.MethodPost:
		a.handleCreatePrivateMessage(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *App) handleListPrivateMessages(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	peerID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if peerID == "" {
		writeError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}
	if peerID == currentUser.ID {
		writeError(w, http.StatusBadRequest, "you cannot open a private chat with yourself")
		return
	}

	if _, err := a.getUserByID(r.Context(), peerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "target user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load target user")
		return
	}

	allowed, err := a.canPrivateChat(r, currentUser.ID, peerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check private chat permissions")
		return
	}
	if !allowed {
		writeError(w, http.StatusForbidden, "private chat requires users to follow each other")
		return
	}

	messages, err := a.loadPrivateMessagesBetween(r, currentUser.ID, peerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load private messages")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]chatMessageItem{"messages": messages})
}

func (a *App) handleCreatePrivateMessage(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req privateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.RecipientID = strings.TrimSpace(req.RecipientID)
	req.Content = strings.TrimSpace(req.Content)
	if req.RecipientID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "recipient_id and content are required")
		return
	}
	if req.RecipientID == currentUser.ID {
		writeError(w, http.StatusBadRequest, "you cannot send a private message to yourself")
		return
	}

	if _, err := a.getUserByID(r.Context(), req.RecipientID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "recipient not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load recipient")
		return
	}

	allowed, err := a.canPrivateChat(r, currentUser.ID, req.RecipientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check private chat permissions")
		return
	}
	if !allowed {
		writeError(w, http.StatusForbidden, "private chat requires users to follow each other")
		return
	}

	messageID := uuid.NewString()
	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO private_messages (id, sender_id, recipient_id, content, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, messageID, currentUser.ID, req.RecipientID, req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create private message")
		return
	}

	message, err := a.loadPrivateMessageByID(r, messageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "private message created but failed to load it")
		return
	}

	a.wsHub.sendToUser(req.RecipientID, map[string]any{
		"type": "private_message",
		"data": message,
	})

	senderName := strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName)
	if senderName == "" {
		senderName = currentUser.Email
	}

	_ = a.pushNotification(r.Context(), req.RecipientID, "private_message_received", map[string]any{
		"message_id":   message.ID,
		"sender_id":    currentUser.ID,
		"sender_name":  senderName,
		"content":      message.Content,
		"recipient_id": req.RecipientID,
	})

	writeJSON(w, http.StatusCreated, map[string]chatMessageItem{"message": message})
}

func (a *App) handleGroupMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListGroupMessages(w, r)
	case http.MethodPost:
		a.handleCreateGroupMessage(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *App) handleListGroupMessages(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	groupID := strings.TrimSpace(r.URL.Query().Get("group_id"))
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "group_id query parameter is required")
		return
	}

	isMember, _, err := a.groupMembership(r, groupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check group membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to access group chat")
		return
	}

	messages, err := a.loadGroupMessages(r, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load group messages")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]chatMessageItem{"messages": messages})
}

func (a *App) handleCreateGroupMessage(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req groupMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.GroupID = strings.TrimSpace(req.GroupID)
	req.Content = strings.TrimSpace(req.Content)
	if req.GroupID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "group_id and content are required")
		return
	}

	isMember, _, err := a.groupMembership(r, req.GroupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check group membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to send group messages")
		return
	}

	messageID := uuid.NewString()
	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO group_messages (id, group_id, sender_id, content, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, messageID, req.GroupID, currentUser.ID, req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create group message")
		return
	}

	message, err := a.loadGroupMessageByID(r, messageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "group message created but failed to load it")
		return
	}

	memberIDs, err := a.loadGroupMemberIDs(r, req.GroupID)
	if err == nil {
		senderName := strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName)
		if senderName == "" {
			senderName = currentUser.Email
		}

		groupTitle := ""
		_ = a.db.QueryRowContext(r.Context(), `
			SELECT title
			FROM groups
			WHERE id = ?
			LIMIT 1
		`, req.GroupID).Scan(&groupTitle)

		targets := make([]string, 0, len(memberIDs))
		for _, memberID := range memberIDs {
			if memberID != currentUser.ID {
				targets = append(targets, memberID)
			}
		}
		a.wsHub.sendToUsers(targets, map[string]any{
			"type": "group_message",
			"data": message,
		})

		for _, targetID := range targets {
			_ = a.pushNotification(r.Context(), targetID, "group_message_received", map[string]any{
				"message_id":  message.ID,
				"group_id":    req.GroupID,
				"group_title": groupTitle,
				"sender_id":   currentUser.ID,
				"sender_name": senderName,
				"content":     message.Content,
			})
		}
	}

	writeJSON(w, http.StatusCreated, map[string]chatMessageItem{"message": message})
}

func (a *App) canPrivateChat(r *http.Request, userA, userB string) (bool, error) {
	if userA == userB {
		return false, nil
	}

	// The subject requires at least ONE of the users to be following the other
	// ("users that they are following or being followed").
	aFollowsB, err := a.isFollowing(r, userA, userB)
	if err != nil {
		return false, err
	}
	if aFollowsB {
		return true, nil
	}

	bFollowsA, err := a.isFollowing(r, userB, userA)
	if err != nil {
		return false, err
	}

	return bFollowsA, nil
}

func (a *App) loadPrivateMessagesBetween(r *http.Request, userA, userB string) ([]chatMessageItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			pm.id,
			pm.sender_id,
			pm.recipient_id,
			pm.content,
			CAST(pm.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM private_messages pm
		JOIN users u ON u.id = pm.sender_id
		WHERE (pm.sender_id = ? AND pm.recipient_id = ?)
			OR (pm.sender_id = ? AND pm.recipient_id = ?)
		ORDER BY pm.created_at ASC, pm.id ASC
	`, userA, userB, userB, userA)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]chatMessageItem, 0)
	for rows.Next() {
		var message chatMessageItem
		var recipientID string
		var avatar sql.NullString
		var nickname sql.NullString

		if err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&recipientID,
			&message.Content,
			&message.CreatedAt,
			&message.Sender.ID,
			&message.Sender.FirstName,
			&message.Sender.LastName,
			&avatar,
			&nickname,
		); err != nil {
			return nil, err
		}

		message.RecipientID = &recipientID
		message.GroupID = nil
		message.Sender.AvatarPath = ptrFromNull(avatar)
		message.Sender.Nickname = ptrFromNull(nickname)
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (a *App) loadPrivateMessageByID(r *http.Request, messageID string) (chatMessageItem, error) {
	var message chatMessageItem
	var recipientID string
	var avatar sql.NullString
	var nickname sql.NullString

	err := a.db.QueryRowContext(r.Context(), `
		SELECT
			pm.id,
			pm.sender_id,
			pm.recipient_id,
			pm.content,
			CAST(pm.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM private_messages pm
		JOIN users u ON u.id = pm.sender_id
		WHERE pm.id = ?
		LIMIT 1
	`, messageID).Scan(
		&message.ID,
		&message.SenderID,
		&recipientID,
		&message.Content,
		&message.CreatedAt,
		&message.Sender.ID,
		&message.Sender.FirstName,
		&message.Sender.LastName,
		&avatar,
		&nickname,
	)
	if err != nil {
		return chatMessageItem{}, err
	}

	message.RecipientID = &recipientID
	message.GroupID = nil
	message.Sender.AvatarPath = ptrFromNull(avatar)
	message.Sender.Nickname = ptrFromNull(nickname)
	return message, nil
}

func (a *App) loadGroupMessages(r *http.Request, groupID string) ([]chatMessageItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			gm.id,
			gm.group_id,
			gm.sender_id,
			gm.content,
			CAST(gm.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM group_messages gm
		JOIN users u ON u.id = gm.sender_id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at ASC, gm.id ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]chatMessageItem, 0)
	for rows.Next() {
		var message chatMessageItem
		var groupIDValue string
		var avatar sql.NullString
		var nickname sql.NullString

		if err := rows.Scan(
			&message.ID,
			&groupIDValue,
			&message.SenderID,
			&message.Content,
			&message.CreatedAt,
			&message.Sender.ID,
			&message.Sender.FirstName,
			&message.Sender.LastName,
			&avatar,
			&nickname,
		); err != nil {
			return nil, err
		}

		message.GroupID = &groupIDValue
		message.RecipientID = nil
		message.Sender.AvatarPath = ptrFromNull(avatar)
		message.Sender.Nickname = ptrFromNull(nickname)
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (a *App) loadGroupMessageByID(r *http.Request, messageID string) (chatMessageItem, error) {
	var message chatMessageItem
	var groupIDValue string
	var avatar sql.NullString
	var nickname sql.NullString

	err := a.db.QueryRowContext(r.Context(), `
		SELECT
			gm.id,
			gm.group_id,
			gm.sender_id,
			gm.content,
			CAST(gm.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM group_messages gm
		JOIN users u ON u.id = gm.sender_id
		WHERE gm.id = ?
		LIMIT 1
	`, messageID).Scan(
		&message.ID,
		&groupIDValue,
		&message.SenderID,
		&message.Content,
		&message.CreatedAt,
		&message.Sender.ID,
		&message.Sender.FirstName,
		&message.Sender.LastName,
		&avatar,
		&nickname,
	)
	if err != nil {
		return chatMessageItem{}, err
	}

	message.GroupID = &groupIDValue
	message.RecipientID = nil
	message.Sender.AvatarPath = ptrFromNull(avatar)
	message.Sender.Nickname = ptrFromNull(nickname)
	return message, nil
}
