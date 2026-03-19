package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type notificationItem struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	IsRead    bool           `json:"is_read"`
	CreatedAt string         `json:"created_at"`
}

type markNotificationsReadRequest struct {
	NotificationID string `json:"notification_id"`
	ReadAll        bool   `json:"read_all"`
}

func (a *App) handleNotifications(w http.ResponseWriter, r *http.Request) {
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

	notifications, err := a.loadNotifications(r, currentUser.ID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load notifications")
		return
	}

	unreadCount, err := a.countUnreadNotifications(r, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load notifications count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

func (a *App) handleMarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
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

	var req markNotificationsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.NotificationID = strings.TrimSpace(req.NotificationID)

	var result sql.Result
	if req.ReadAll {
		result, err = a.db.ExecContext(r.Context(), `
			UPDATE notifications
			SET is_read = 1
			WHERE user_id = ?
				AND is_read = 0
		`, currentUser.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to mark notifications as read")
			return
		}
	} else {
		if req.NotificationID == "" {
			writeError(w, http.StatusBadRequest, "notification_id is required when read_all is false")
			return
		}

		result, err = a.db.ExecContext(r.Context(), `
			UPDATE notifications
			SET is_read = 1
			WHERE id = ?
				AND user_id = ?
		`, req.NotificationID, currentUser.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to mark notification as read")
			return
		}
	}

	affected, _ := result.RowsAffected()
	unreadCount, err := a.countUnreadNotifications(r, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load notifications count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"updated":      affected,
		"unread_count": unreadCount,
	})
}

func (a *App) loadNotifications(r *http.Request, userID string, limit int) ([]notificationItem, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			id,
			user_id,
			type,
			payload,
			is_read,
			CAST(created_at AS TEXT)
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]notificationItem, 0)
	for rows.Next() {
		var item notificationItem
		var payloadRaw sql.NullString
		var isRead int

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Type,
			&payloadRaw,
			&isRead,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		item.IsRead = isRead == 1
		if payloadRaw.Valid && strings.TrimSpace(payloadRaw.String) != "" {
			parsed := make(map[string]any)
			if err := json.Unmarshal([]byte(payloadRaw.String), &parsed); err == nil {
				item.Payload = parsed
			}
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (a *App) countUnreadNotifications(r *http.Request, userID string) (int, error) {
	var count int
	err := a.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ?
			AND is_read = 0
	`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *App) createNotification(ctx context.Context, userID, notifType string, payload map[string]any) (notificationItem, error) {
	item := notificationItem{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      notifType,
		IsRead:    false,
		CreatedAt: a.requestNowISO(),
	}

	var payloadJSON any
	if len(payload) > 0 {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return notificationItem{}, err
		}
		payloadJSON = string(encoded)
		item.Payload = payload
	}

	_, err := a.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, payload, is_read, created_at)
		VALUES (?, ?, ?, ?, 0, CURRENT_TIMESTAMP)
	`, item.ID, userID, notifType, payloadJSON)
	if err != nil {
		return notificationItem{}, err
	}

	return item, nil
}

func (a *App) pushNotification(ctx context.Context, userID, notifType string, payload map[string]any) error {
	item, err := a.createNotification(ctx, userID, notifType, payload)
	if err != nil {
		return err
	}

	a.wsHub.sendToUser(userID, map[string]any{
		"type": "notification",
		"data": item,
	})

	return nil
}
