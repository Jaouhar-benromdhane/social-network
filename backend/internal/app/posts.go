package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const maxPostMediaSize = 8 << 20 // 8MB

type postAuthor struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	AvatarPath *string `json:"avatar_path,omitempty"`
	Nickname   *string `json:"nickname,omitempty"`
}

type commentItem struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	MediaPath *string    `json:"media_path,omitempty"`
	MediaType *string    `json:"media_type,omitempty"`
	CreatedAt string     `json:"created_at"`
	Author    postAuthor `json:"author"`
}

type postItem struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	Content        string        `json:"content"`
	MediaPath      *string       `json:"media_path,omitempty"`
	MediaType      *string       `json:"media_type,omitempty"`
	Privacy        string        `json:"privacy"`
	CreatedAt      string        `json:"created_at"`
	Author         postAuthor    `json:"author"`
	AllowedUserIDs []string      `json:"allowed_user_ids,omitempty"`
	Comments       []commentItem `json:"comments"`
}

func (a *App) handleCreatePost(w http.ResponseWriter, r *http.Request) {
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

	content := strings.TrimSpace(r.FormValue("content"))
	privacy := strings.TrimSpace(r.FormValue("privacy"))
	if privacy == "" {
		privacy = "public"
	}
	if privacy != "public" && privacy != "almost_private" && privacy != "private" {
		writeError(w, http.StatusBadRequest, "privacy must be public, almost_private or private")
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

	allowedUserIDs, err := parseAllowedUserIDs(r.FormValue("allowed_user_ids"))
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusBadRequest, "allowed_user_ids must be a JSON array or comma-separated list")
		return
	}

	if privacy == "private" {
		allowedUserIDs, err = a.keepOnlyFollowers(r, currentUser.ID, allowedUserIDs)
		if err != nil {
			a.deleteUploadedMedia(mediaPath)
			writeError(w, http.StatusInternalServerError, "failed to validate allowed users")
			return
		}
		if len(allowedUserIDs) == 0 {
			a.deleteUploadedMedia(mediaPath)
			writeError(w, http.StatusBadRequest, "private posts require at least one allowed follower")
			return
		}
	} else {
		allowedUserIDs = nil
	}

	postID := uuid.NewString()
	var mediaTypeValue any
	if mediaType != nil {
		mediaTypeValue = *mediaType
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO posts (id, user_id, content, media_path, media_type, privacy, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, postID, currentUser.ID, content, nullStringInput(mediaPath), mediaTypeValue, privacy)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	for _, allowedUserID := range allowedUserIDs {
		_, err = tx.ExecContext(r.Context(), `
			INSERT OR IGNORE INTO post_allowed_users (post_id, user_id)
			VALUES (?, ?)
		`, postID, allowedUserID)
		if err != nil {
			a.deleteUploadedMedia(mediaPath)
			writeError(w, http.StatusInternalServerError, "failed to save allowed users")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to commit post")
		return
	}

	post, err := a.loadPostByIDForViewer(r, currentUser.ID, postID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "post created but failed to load it")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]postItem{"post": post})
}

func (a *App) handleCreateComment(w http.ResponseWriter, r *http.Request) {
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

	postID := strings.TrimSpace(r.FormValue("post_id"))
	content := strings.TrimSpace(r.FormValue("content"))
	if postID == "" {
		writeError(w, http.StatusBadRequest, "post_id is required")
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

	ownerID, privacy, err := a.loadPostOwnerAndPrivacy(r, postID)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load post")
		return
	}

	canView, err := a.canViewPost(r, currentUser.ID, ownerID, privacy, postID)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to check post permissions")
		return
	}
	if !canView {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusForbidden, "you cannot comment this post")
		return
	}

	commentID := uuid.NewString()
	var mediaTypeValue any
	if mediaType != nil {
		mediaTypeValue = *mediaType
	}

	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO comments (id, post_id, user_id, content, media_path, media_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, commentID, postID, currentUser.ID, content, nullStringInput(mediaPath), mediaTypeValue)
	if err != nil {
		a.deleteUploadedMedia(mediaPath)
		writeError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	comment := commentItem{
		ID:        commentID,
		PostID:    postID,
		UserID:    currentUser.ID,
		Content:   content,
		MediaPath: nullOrStringPtr(mediaPath),
		MediaType: mediaType,
		CreatedAt: a.requestNowISO(),
		Author: postAuthor{
			ID:         currentUser.ID,
			FirstName:  currentUser.FirstName,
			LastName:   currentUser.LastName,
			AvatarPath: currentUser.AvatarPath,
			Nickname:   currentUser.Nickname,
		},
	}

	writeJSON(w, http.StatusCreated, map[string]commentItem{"comment": comment})
}

func (a *App) handleFeedPosts(w http.ResponseWriter, r *http.Request) {
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

	targetUserID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	var ownerID *string
	if targetUserID != "" {
		ownerID = &targetUserID
	}

	posts, err := a.loadVisiblePosts(r, currentUser.ID, ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]postItem{"posts": posts})
}

func (a *App) loadVisiblePosts(r *http.Request, viewerID string, ownerID *string) ([]postItem, error) {
	query := `
		SELECT
			p.id,
			p.user_id,
			p.content,
			p.media_path,
			p.media_type,
			p.privacy,
			CAST(p.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM posts p
		JOIN users u ON u.id = p.user_id
	`
	args := make([]any, 0, 1)
	if ownerID != nil {
		query += " WHERE p.user_id = ?"
		args = append(args, *ownerID)
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := a.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]postItem, 0)
	for rows.Next() {
		var post postItem
		var postMediaPath sql.NullString
		var postMediaType sql.NullString
		var authorAvatar sql.NullString
		var authorNickname sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&postMediaPath,
			&postMediaType,
			&post.Privacy,
			&post.CreatedAt,
			&post.Author.ID,
			&post.Author.FirstName,
			&post.Author.LastName,
			&authorAvatar,
			&authorNickname,
		)
		if err != nil {
			return nil, err
		}

		post.MediaPath = ptrFromNull(postMediaPath)
		post.MediaType = ptrFromNull(postMediaType)
		post.Author.AvatarPath = ptrFromNull(authorAvatar)
		post.Author.Nickname = ptrFromNull(authorNickname)

		canView, err := a.canViewPost(r, viewerID, post.UserID, post.Privacy, post.ID)
		if err != nil {
			return nil, err
		}
		if !canView {
			continue
		}

		if viewerID == post.UserID {
			allowedUserIDs, err := a.loadAllowedUsersForPost(r, post.ID)
			if err != nil {
				return nil, err
			}
			post.AllowedUserIDs = allowedUserIDs
		}

		comments, err := a.loadCommentsForPost(r, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (a *App) loadPostByIDForViewer(r *http.Request, viewerID, postID string) (postItem, error) {
	posts, err := a.loadVisiblePosts(r, viewerID, nil)
	if err != nil {
		return postItem{}, err
	}

	for _, post := range posts {
		if post.ID == postID {
			return post, nil
		}
	}

	return postItem{}, sql.ErrNoRows
}

func (a *App) loadPostOwnerAndPrivacy(r *http.Request, postID string) (string, string, error) {
	var ownerID string
	var privacy string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT user_id, privacy
		FROM posts
		WHERE id = ?
		LIMIT 1
	`, postID).Scan(&ownerID, &privacy)
	if err != nil {
		return "", "", err
	}

	return ownerID, privacy, nil
}

func (a *App) canViewPost(r *http.Request, viewerID, ownerID, privacy, postID string) (bool, error) {
	if viewerID == ownerID {
		return true, nil
	}

	switch privacy {
	case "public":
		return true, nil
	case "almost_private":
		return a.isFollowing(r, viewerID, ownerID)
	case "private":
		return a.isPrivatePostAllowed(r, postID, viewerID)
	default:
		return false, nil
	}
}

func (a *App) isPrivatePostAllowed(r *http.Request, postID, viewerID string) (bool, error) {
	var allowed bool
	err := a.db.QueryRowContext(r.Context(), `
		SELECT EXISTS(
			SELECT 1
			FROM post_allowed_users
			WHERE post_id = ?
				AND user_id = ?
		)
	`, postID, viewerID).Scan(&allowed)
	if err != nil {
		return false, err
	}

	return allowed, nil
}

func (a *App) loadAllowedUsersForPost(r *http.Request, postID string) ([]string, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT user_id
		FROM post_allowed_users
		WHERE post_id = ?
		ORDER BY user_id ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allowed := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		allowed = append(allowed, userID)
	}

	return allowed, rows.Err()
}

func (a *App) loadCommentsForPost(r *http.Request, postID string) ([]commentItem, error) {
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			c.content,
			c.media_path,
			c.media_type,
			CAST(c.created_at AS TEXT),
			u.id,
			u.first_name,
			u.last_name,
			u.avatar_path,
			u.nickname
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]commentItem, 0)
	for rows.Next() {
		var comment commentItem
		var commentMediaPath sql.NullString
		var commentMediaType sql.NullString
		var authorAvatar sql.NullString
		var authorNickname sql.NullString

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&commentMediaPath,
			&commentMediaType,
			&comment.CreatedAt,
			&comment.Author.ID,
			&comment.Author.FirstName,
			&comment.Author.LastName,
			&authorAvatar,
			&authorNickname,
		)
		if err != nil {
			return nil, err
		}

		comment.MediaPath = ptrFromNull(commentMediaPath)
		comment.MediaType = ptrFromNull(commentMediaType)
		comment.Author.AvatarPath = ptrFromNull(authorAvatar)
		comment.Author.Nickname = ptrFromNull(authorNickname)
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func parseAllowedUserIDs(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}, nil
	}

	values := make([]string, 0)
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return nil, err
		}
	} else {
		values = strings.Split(raw, ",")
	}

	unique := make(map[string]struct{}, len(values))
	clean := make([]string, 0, len(values))
	for _, value := range values {
		id := strings.TrimSpace(value)
		if id == "" {
			continue
		}
		if _, exists := unique[id]; exists {
			continue
		}
		unique[id] = struct{}{}
		clean = append(clean, id)
	}

	return clean, nil
}

func (a *App) keepOnlyFollowers(r *http.Request, ownerID string, candidateIDs []string) ([]string, error) {
	if len(candidateIDs) == 0 {
		return []string{}, nil
	}

	query := fmt.Sprintf(`
		SELECT follower_id
		FROM follows
		WHERE following_id = ?
			AND follower_id IN (%s)
	`, placeholders(len(candidateIDs)))

	args := make([]any, 0, 1+len(candidateIDs))
	args = append(args, ownerID)
	for _, id := range candidateIDs {
		args = append(args, id)
	}

	rows, err := a.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	valid := make(map[string]struct{}, len(candidateIDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		valid[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		if _, ok := valid[id]; ok {
			filtered = append(filtered, id)
		}
	}

	return filtered, nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func (a *App) savePostMediaFromRequest(r *http.Request, fieldName, subDir string, maxSize int64) (string, *string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("invalid media upload")
	}
	defer file.Close()

	if header.Size > maxSize {
		return "", nil, fmt.Errorf("media exceeds 8MB")
	}

	head := make([]byte, 512)
	n, readErr := file.Read(head)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", nil, fmt.Errorf("failed reading media")
	}

	contentType := http.DetectContentType(head[:n])
	var ext string
	var mediaType string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
		mediaType = "image"
	case "image/png":
		ext = ".png"
		mediaType = "image"
	case "image/gif":
		ext = ".gif"
		mediaType = "gif"
	default:
		return "", nil, fmt.Errorf("media must be JPEG, PNG or GIF")
	}

	fileName := uuid.NewString() + ext
	relativePath := filepath.ToSlash(filepath.Join(subDir, fileName))
	absolutePath := filepath.Join(a.uploadDir, filepath.FromSlash(relativePath))

	output, err := os.Create(absolutePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed saving media")
	}
	defer output.Close()

	if _, err := output.Write(head[:n]); err != nil {
		return "", nil, fmt.Errorf("failed writing media")
	}
	if _, err := io.Copy(output, file); err != nil {
		return "", nil, fmt.Errorf("failed writing media")
	}

	publicPath := "/uploads/" + relativePath
	return publicPath, &mediaType, nil
}

func (a *App) deleteUploadedMedia(publicPath string) {
	if publicPath == "" || !strings.HasPrefix(publicPath, "/uploads/") {
		return
	}

	relative := strings.TrimPrefix(publicPath, "/uploads/")
	absolute := filepath.Join(a.uploadDir, filepath.FromSlash(relative))
	_ = os.Remove(absolute)
}

func nullOrStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
