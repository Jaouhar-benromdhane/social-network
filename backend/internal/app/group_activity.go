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

type groupPostItem struct {
	ID        string             `json:"id"`
	GroupID   string             `json:"group_id"`
	UserID    string             `json:"user_id"`
	Content   string             `json:"content"`
	MediaPath *string            `json:"media_path,omitempty"`
	MediaType *string            `json:"media_type,omitempty"`
	CreatedAt string             `json:"created_at"`
	Author    postAuthor         `json:"author"`
	Comments  []groupCommentItem `json:"comments"`
}

type groupCommentItem struct {
	ID          string     `json:"id"`
	GroupPostID string     `json:"group_post_id"`
	UserID      string     `json:"user_id"`
	Content     string     `json:"content"`
	MediaPath   *string    `json:"media_path,omitempty"`
	MediaType   *string    `json:"media_type,omitempty"`
	CreatedAt   string     `json:"created_at"`
	Author      postAuthor `json:"author"`
}

type createGroupEventRequest struct {
	GroupID       string   `json:"group_id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	EventDateTime string   `json:"event_datetime"`
	Options       []string `json:"options"`
}

type groupEventOptionItem struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	Label     string `json:"label"`
	VoteCount int    `json:"vote_count"`
}

type groupEventItem struct {
	ID             string                 `json:"id"`
	GroupID        string                 `json:"group_id"`
	CreatorID      string                 `json:"creator_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	EventDateTime  string                 `json:"event_datetime"`
	CreatedAt      string                 `json:"created_at"`
	Options        []groupEventOptionItem `json:"options"`
	MyVoteOptionID *string                `json:"my_vote_option_id,omitempty"`
}

type voteGroupEventRequest struct {
	EventID  string `json:"event_id"`
	OptionID string `json:"option_id"`
}

func (a *App) handleGroupPosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListGroupPosts(w, r)
	case http.MethodPost:
		a.handleCreateGroupPost(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *App) handleCreateGroupPost(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	if err := r.ParseMultipartForm(maxPostMediaSize + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	groupID := strings.TrimSpace(r.FormValue("group_id"))
	content := strings.TrimSpace(r.FormValue("content"))
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "group_id is required")
		return
	}

	isMember, _, err := a.groupMembership(r, groupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to create posts")
		return
	}

	mediaPath, mediaType, err := a.savePostMediaFromRequest(r, "media", "posts", maxPostMediaSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if content == "" && mediaPath == "" {
		writeError(w, http.StatusBadRequest, "content or media is required")
		return
	}

	groupPostID := uuid.NewString()
	var mediaTypeValue any
	if mediaType != nil {
		mediaTypeValue = *mediaType
	}

	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO group_posts (id, group_id, user_id, content, media_path, media_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, groupPostID, groupID, currentUser.ID, content, nullStringInput(mediaPath), mediaTypeValue)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to create group post")
		return
	}

	post, err := a.loadGroupPostByID(r, groupID, groupPostID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "group post created but failed to load it")
		return
	}

	if memberIDs, err := a.loadGroupMemberIDs(r, groupID); err == nil {
		a.pushRealtimeEventToUsers(memberIDs, "groups_updated", map[string]any{
			"reason":        "group_post_created",
			"group_id":      groupID,
			"group_post_id": groupPostID,
			"actor_id":      currentUser.ID,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]groupPostItem{"post": post})
}

func (a *App) handleListGroupPosts(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to view group posts")
		return
	}

	posts, err := a.loadGroupPosts(r, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load group posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]groupPostItem{"posts": posts})
}

func (a *App) handleCreateGroupComment(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseMultipartForm(maxPostMediaSize + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	groupPostID := strings.TrimSpace(r.FormValue("group_post_id"))
	content := strings.TrimSpace(r.FormValue("content"))
	if groupPostID == "" {
		writeError(w, http.StatusBadRequest, "group_post_id is required")
		return
	}

	mediaPath, mediaType, err := a.savePostMediaFromRequest(r, "media", "comments", maxPostMediaSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if content == "" && mediaPath == "" {
		writeError(w, http.StatusBadRequest, "content or media is required")
		return
	}

	var groupID string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT group_id
		FROM group_posts
		WHERE id = ?
		LIMIT 1
	`, groupPostID).Scan(&groupID)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load group post")
		return
	}

	isMember, _, err := a.groupMembership(r, groupID, currentUser.ID)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusForbidden, "you must be a group member to comment")
		return
	}

	groupCommentID := uuid.NewString()
	var mediaTypeValue any
	if mediaType != nil {
		mediaTypeValue = *mediaType
	}

	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO group_comments (id, group_post_id, user_id, content, media_path, media_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, groupCommentID, groupPostID, currentUser.ID, content, nullStringInput(mediaPath), mediaTypeValue)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to create group comment")
		return
	}

	comment := groupCommentItem{
		ID:          groupCommentID,
		GroupPostID: groupPostID,
		UserID:      currentUser.ID,
		Content:     content,
		MediaPath:   nullOrStringPtr(mediaPath),
		MediaType:   mediaType,
		CreatedAt:   a.requestNowISO(),
		Author: postAuthor{
			ID:         currentUser.ID,
			FirstName:  currentUser.FirstName,
			LastName:   currentUser.LastName,
			AvatarPath: currentUser.AvatarPath,
			Nickname:   currentUser.Nickname,
		},
	}

	if memberIDs, err := a.loadGroupMemberIDs(r, groupID); err == nil {
		a.pushRealtimeEventToUsers(memberIDs, "groups_updated", map[string]any{
			"reason":           "group_post_commented",
			"group_id":         groupID,
			"group_post_id":    groupPostID,
			"group_comment_id": groupCommentID,
			"actor_id":         currentUser.ID,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]groupCommentItem{"comment": comment})
}

func (a *App) handleGroupEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListGroupEvents(w, r)
	case http.MethodPost:
		a.handleCreateGroupEvent(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *App) handleCreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	var req createGroupEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.GroupID = strings.TrimSpace(req.GroupID)
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.EventDateTime = strings.TrimSpace(req.EventDateTime)
	if req.GroupID == "" || req.Title == "" || req.Description == "" || req.EventDateTime == "" {
		writeError(w, http.StatusBadRequest, "group_id, title, description and event_datetime are required")
		return
	}

	cleanOptions := make([]string, 0, len(req.Options))
	for _, option := range req.Options {
		trimmed := strings.TrimSpace(option)
		if trimmed != "" {
			cleanOptions = append(cleanOptions, trimmed)
		}
	}
	if len(cleanOptions) < 2 {
		writeError(w, http.StatusBadRequest, "at least two event options are required")
		return
	}

	eventTime, err := parseEventDateTime(req.EventDateTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "event_datetime must be RFC3339 or YYYY-MM-DDTHH:MM")
		return
	}

	isMember, _, err := a.groupMembership(r, req.GroupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to create events")
		return
	}

	eventID := uuid.NewString()
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO events (id, group_id, creator_id, title, description, event_datetime, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, eventID, req.GroupID, currentUser.ID, req.Title, req.Description, eventTime.UTC().Format(time.RFC3339))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	for _, option := range cleanOptions {
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO event_options (id, event_id, label)
			VALUES (?, ?, ?)
		`, uuid.NewString(), eventID, option)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create event options")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit event creation")
		return
	}

	event, err := a.loadGroupEventByID(r, currentUser.ID, eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "event created but failed to load it")
		return
	}

	memberIDs, err := a.loadGroupMemberIDs(r, req.GroupID)
	if err == nil {
		groupTitle := ""
		_ = a.db.QueryRowContext(r.Context(), `
			SELECT title
			FROM groups
			WHERE id = ?
			LIMIT 1
		`, req.GroupID).Scan(&groupTitle)

		for _, memberID := range memberIDs {
			if memberID == currentUser.ID {
				continue
			}
			_ = a.pushNotification(r.Context(), memberID, "group_event_created", map[string]any{
				"event_id":         event.ID,
				"group_id":         req.GroupID,
				"group_title":      groupTitle,
				"event_title":      event.Title,
				"event_datetime":   event.EventDateTime,
				"creator_id":       currentUser.ID,
				"creator_name":     strings.TrimSpace(currentUser.FirstName + " " + currentUser.LastName),
				"creator_nickname": currentUser.Nickname,
			})
		}

		a.pushRealtimeEventToUsers(memberIDs, "groups_updated", map[string]any{
			"reason":      "group_event_created",
			"group_id":    req.GroupID,
			"event_id":    event.ID,
			"actor_id":    currentUser.ID,
			"event_title": event.Title,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]groupEventItem{"event": event})
}

func (a *App) handleListGroupEvents(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to view events")
		return
	}

	events, err := a.loadGroupEvents(r, currentUser.ID, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]groupEventItem{"events": events})
}

func (a *App) handleVoteGroupEvent(w http.ResponseWriter, r *http.Request) {
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

	var req voteGroupEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.EventID = strings.TrimSpace(req.EventID)
	req.OptionID = strings.TrimSpace(req.OptionID)
	if req.EventID == "" || req.OptionID == "" {
		writeError(w, http.StatusBadRequest, "event_id and option_id are required")
		return
	}

	var groupID string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT e.group_id
		FROM events e
		JOIN event_options eo ON eo.event_id = e.id
		WHERE e.id = ?
			AND eo.id = ?
		LIMIT 1
	`, req.EventID, req.OptionID).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "event or option not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load event option")
		return
	}

	isMember, _, err := a.groupMembership(r, groupID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "you must be a group member to vote")
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		DELETE FROM event_votes
		WHERE event_id = ?
			AND user_id = ?
	`, req.EventID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset previous vote")
		return
	}

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO event_votes (event_id, event_option_id, user_id, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, req.EventID, req.OptionID, currentUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store vote")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit vote")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "voted",
		"event_id":  req.EventID,
		"option_id": req.OptionID,
	})
}

func (a *App) loadGroupPosts(r *http.Request, groupID string) ([]groupPostItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			gp.id,
			gp.group_id,
			gp.user_id,
			gp.content,
			gp.media_path,
			gp.media_type,
			CAST(gp.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM group_posts gp
		JOIN users u ON u.id = gp.user_id
		WHERE gp.group_id = ?
		ORDER BY gp.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]groupPostItem, 0)
	for rows.Next() {
		var post groupPostItem
		var mediaPath sql.NullString
		var mediaType sql.NullString
		var authorAvatar sql.NullString
		var authorNickname sql.NullString

		if err := rows.Scan(
			&post.ID,
			&post.GroupID,
			&post.UserID,
			&post.Content,
			&mediaPath,
			&mediaType,
			&post.CreatedAt,
			&post.Author.ID,
			&post.Author.FirstName,
			&post.Author.LastName,
			&authorAvatar,
			&authorNickname,
		); err != nil {
			return nil, err
		}

		post.MediaPath = ptrFromNull(mediaPath)
		post.MediaType = ptrFromNull(mediaType)
		post.Author.AvatarPath = ptrFromNull(authorAvatar)
		post.Author.Nickname = ptrFromNull(authorNickname)

		comments, err := a.loadGroupCommentsForPost(r, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (a *App) loadGroupPostByID(r *http.Request, groupID, groupPostID string) (groupPostItem, error) {
	posts, err := a.loadGroupPosts(r, groupID)
	if err != nil {
		return groupPostItem{}, err
	}

	for _, post := range posts {
		if post.ID == groupPostID {
			return post, nil
		}
	}

	return groupPostItem{}, sql.ErrNoRows
}

func (a *App) loadGroupCommentsForPost(r *http.Request, groupPostID string) ([]groupCommentItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			gc.id,
			gc.group_post_id,
			gc.user_id,
			gc.content,
			gc.media_path,
			gc.media_type,
			CAST(gc.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM group_comments gc
		JOIN users u ON u.id = gc.user_id
		WHERE gc.group_post_id = ?
		ORDER BY gc.created_at ASC
	`, groupPostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]groupCommentItem, 0)
	for rows.Next() {
		var comment groupCommentItem
		var mediaPath sql.NullString
		var mediaType sql.NullString
		var authorAvatar sql.NullString
		var authorNickname sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.GroupPostID,
			&comment.UserID,
			&comment.Content,
			&mediaPath,
			&mediaType,
			&comment.CreatedAt,
			&comment.Author.ID,
			&comment.Author.FirstName,
			&comment.Author.LastName,
			&authorAvatar,
			&authorNickname,
		); err != nil {
			return nil, err
		}

		comment.MediaPath = ptrFromNull(mediaPath)
		comment.MediaType = ptrFromNull(mediaType)
		comment.Author.AvatarPath = ptrFromNull(authorAvatar)
		comment.Author.Nickname = ptrFromNull(authorNickname)
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func (a *App) loadGroupEvents(r *http.Request, viewerID, groupID string) ([]groupEventItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			e.id,
			e.group_id,
			e.creator_id,
			e.title,
			e.description,
			CAST(e.event_datetime AS TEXT),
			CAST(e.created_at AS TEXT)
		FROM events e
		WHERE e.group_id = ?
		ORDER BY e.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]groupEventItem, 0)
	for rows.Next() {
		var event groupEventItem
		if err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&event.EventDateTime,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		options, err := a.loadEventOptions(r, event.ID)
		if err != nil {
			return nil, err
		}
		event.Options = options

		myVoteID, err := a.loadMyEventVoteOptionID(r, event.ID, viewerID)
		if err != nil {
			return nil, err
		}
		event.MyVoteOptionID = myVoteID

		events = append(events, event)
	}

	return events, rows.Err()
}

func (a *App) loadGroupEventByID(r *http.Request, viewerID, eventID string) (groupEventItem, error) {
	var groupID string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT group_id
		FROM events
		WHERE id = ?
		LIMIT 1
	`, eventID).Scan(&groupID)
	if err != nil {
		return groupEventItem{}, err
	}

	events, err := a.loadGroupEvents(r, viewerID, groupID)
	if err != nil {
		return groupEventItem{}, err
	}

	for _, event := range events {
		if event.ID == eventID {
			return event, nil
		}
	}

	return groupEventItem{}, sql.ErrNoRows
}

func (a *App) loadEventOptions(r *http.Request, eventID string) ([]groupEventOptionItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			eo.id,
			eo.event_id,
			eo.label,
			(
				SELECT COUNT(*)
				FROM event_votes ev
				WHERE ev.event_option_id = eo.id
			) AS vote_count
		FROM event_options eo
		WHERE eo.event_id = ?
		ORDER BY eo.id ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]groupEventOptionItem, 0)
	for rows.Next() {
		var option groupEventOptionItem
		if err := rows.Scan(&option.ID, &option.EventID, &option.Label, &option.VoteCount); err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	return options, rows.Err()
}

func (a *App) loadMyEventVoteOptionID(r *http.Request, eventID, userID string) (*string, error) {
	var optionID sql.NullString
	err := a.db.QueryRowContext(r.Context(), `
		SELECT event_option_id
		FROM event_votes
		WHERE event_id = ?
			AND user_id = ?
		LIMIT 1
	`, eventID, userID).Scan(&optionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return ptrFromNull(optionID), nil
}

func parseEventDateTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty datetime")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("invalid datetime format")
}
