package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var errUnauthorized = errors.New("unauthorized")

// User is the public-safe representation returned by the API.
type User struct {
	ID                string  `json:"id"`
	Email             string  `json:"email"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	DateOfBirth       string  `json:"date_of_birth"`
	AvatarPath        *string `json:"avatar_path,omitempty"`
	Nickname          *string `json:"nickname,omitempty"`
	AboutMe           *string `json:"about_me,omitempty"`
	ProfileVisibility string  `json:"profile_visibility"`
	CreatedAt         *string `json:"created_at,omitempty"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func nullStringInput(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func ptrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func (a *App) getUserByID(ctx context.Context, id string) (User, error) {
	row := a.db.QueryRowContext(ctx, `
		SELECT
			id,
			email,
			first_name,
			last_name,
			date_of_birth,
			avatar_path,
			nickname,
			about_me,
			profile_visibility,
			CAST(created_at AS TEXT)
		FROM users
		WHERE id = ?
		LIMIT 1
	`, id)

	return scanUser(row)
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (User, error) {
	var user User
	var avatar sql.NullString
	var nickname sql.NullString
	var aboutMe sql.NullString
	var createdAt sql.NullString

	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&avatar,
		&nickname,
		&aboutMe,
		&user.ProfileVisibility,
		&createdAt,
	)
	if err != nil {
		return User{}, err
	}

	user.AvatarPath = ptrFromNull(avatar)
	user.Nickname = ptrFromNull(nickname)
	user.AboutMe = ptrFromNull(aboutMe)
	user.CreatedAt = ptrFromNull(createdAt)

	return user, nil
}
