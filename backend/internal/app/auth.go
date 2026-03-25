package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "session_token"
	maxAvatarSize     = 5 << 20 // 5MB
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	if err := r.ParseMultipartForm(maxAvatarSize + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	email := normalizeEmail(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	firstName := sanitizeInput(r.FormValue("first_name"))
	lastName := sanitizeInput(r.FormValue("last_name"))
	dateOfBirth := strings.TrimSpace(r.FormValue("date_of_birth"))
	nickname := sanitizeInput(r.FormValue("nickname"))
	aboutMe := sanitizeInput(r.FormValue("about_me"))
	visibility := strings.TrimSpace(r.FormValue("profile_visibility"))
	if visibility == "" {
		visibility = "public"
	}

	if email == "" || password == "" || firstName == "" || lastName == "" || dateOfBirth == "" {
		writeError(w, http.StatusBadRequest, "email, password, first_name, last_name and date_of_birth are required")
		return
	}

	if !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if len(password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	if _, err := time.Parse("2006-01-02", dateOfBirth); err != nil {
		writeError(w, http.StatusBadRequest, "date_of_birth must use YYYY-MM-DD format")
		return
	}

	if visibility != "public" && visibility != "private" {
		writeError(w, http.StatusBadRequest, "profile_visibility must be public or private")
		return
	}

	avatarPath, err := a.saveAvatarFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.deleteUploadedAvatar(avatarPath)
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	userID := uuid.NewString()
	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO users (
			id,
			email,
			password_hash,
			first_name,
			last_name,
			date_of_birth,
			avatar_path,
			nickname,
			about_me,
			profile_visibility,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		userID,
		email,
		string(hash),
		firstName,
		lastName,
		dateOfBirth,
		nullStringInput(avatarPath),
		nullStringInput(nickname),
		nullStringInput(aboutMe),
		visibility,
	)
	if err != nil {
		a.deleteUploadedAvatar(avatarPath)
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.nickname") {
			writeError(w, http.StatusConflict, "nickname already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, expiresAt, err := a.createSession(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	a.setSessionCookie(w, token, expiresAt)

	user, err := a.getUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user created but failed to load profile")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]User{"user": user})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var userID string
	var passwordHash string
	err := a.db.QueryRowContext(r.Context(), `
		SELECT id, password_hash
		FROM users
		WHERE email = ?
		LIMIT 1
	`, email).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, expiresAt, err := a.createSession(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	a.setSessionCookie(w, token, expiresAt)

	user, err := a.getUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]User{"user": user})
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		var userID string
		err = a.db.QueryRowContext(r.Context(), `
			SELECT user_id
			FROM sessions
			WHERE session_token = ?
			LIMIT 1
		`, cookie.Value).Scan(&userID)
		if err == nil && strings.TrimSpace(userID) != "" {
			_ = a.deleteSessionsByUserID(r.Context(), userID)
		} else {
			_ = a.deleteSessionByToken(r.Context(), cookie.Value)
		}
	}
	a.clearSessionCookie(w)

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]User{"user": user})
}

func (a *App) userFromRequest(ctx context.Context, r *http.Request) (User, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return User{}, errUnauthorized
	}

	row := a.db.QueryRowContext(ctx, `
		SELECT
			u.id,
			u.email,
			u.first_name,
			u.last_name,
			u.date_of_birth,
			u.avatar_path,
			u.nickname,
			u.about_me,
			u.profile_visibility,
			CAST(u.created_at AS TEXT)
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.session_token = ?
			AND datetime(s.expires_at) > datetime('now')
		LIMIT 1
	`, cookie.Value)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, errUnauthorized
		}
		return User{}, err
	}

	return user, nil
}

func (a *App) createSession(ctx context.Context, userID string) (string, time.Time, error) {
	token := uuid.NewString()
	expiresAt := time.Now().UTC().Add(a.sessionDuration)

	_, err := a.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, session_token, expires_at, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, uuid.NewString(), userID, token, expiresAt.Format(time.RFC3339))
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (a *App) deleteSessionByToken(ctx context.Context, token string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM sessions WHERE session_token = ?`, token)
	return err
}

func (a *App) deleteSessionsByUserID(ctx context.Context, userID string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (a *App) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   maxAge,
	})
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (a *App) saveAvatarFromRequest(r *http.Request) (string, error) {
	file, header, err := r.FormFile("avatar")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		return "", fmt.Errorf("invalid avatar upload")
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		return "", fmt.Errorf("avatar exceeds 5MB")
	}

	head := make([]byte, 512)
	n, readErr := file.Read(head)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", fmt.Errorf("failed reading avatar")
	}

	contentType := http.DetectContentType(head[:n])
	ext := ""
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return "", fmt.Errorf("avatar must be JPEG, PNG or GIF")
	}

	fileName := uuid.NewString() + ext
	relativeAvatarPath := filepath.ToSlash(filepath.Join("avatars", fileName))
	absoluteAvatarPath := filepath.Join(a.uploadDir, filepath.FromSlash(relativeAvatarPath))

	output, err := os.Create(absoluteAvatarPath)
	if err != nil {
		return "", fmt.Errorf("failed saving avatar")
	}
	defer output.Close()

	if _, err := output.Write(head[:n]); err != nil {
		return "", fmt.Errorf("failed writing avatar")
	}
	if _, err := io.Copy(output, file); err != nil {
		return "", fmt.Errorf("failed writing avatar")
	}

	return "/uploads/" + relativeAvatarPath, nil
}

func (a *App) deleteUploadedAvatar(publicPath string) {
	if publicPath == "" || !strings.HasPrefix(publicPath, "/uploads/") {
		return
	}

	relative := strings.TrimPrefix(publicPath, "/uploads/")
	absolute := filepath.Join(a.uploadDir, filepath.FromSlash(relative))
	_ = os.Remove(absolute)
}
