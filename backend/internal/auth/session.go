package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

const CookieName = "gm_session"

type Store struct {
	Pool *pgxpool.Pool
	TTL  time.Duration
}

func (s *Store) Issue(ctx context.Context, userID string) (*model.Session, error) {
	sess := &model.Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		CSRF:      randomHex(16),
		ExpiresAt: clock.Now().Add(s.TTL),
		CreatedAt: clock.Now(),
	}
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, csrf_token, expires_at, created_at) VALUES ($1,$2,$3,$4,$5)`,
		sess.ID, sess.UserID, sess.CSRF, sess.ExpiresAt, sess.CreatedAt)
	return sess, err
}

func (s *Store) Get(ctx context.Context, id string) (*model.Session, *model.User, error) {
	row := s.Pool.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.csrf_token, s.expires_at, s.created_at,
		       u.id, u.email, u.password_hash, u.display_name, u.role, u.avatar_url, u.bio, u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.id = $1`, id)
	var sess model.Session
	var u model.User
	if err := row.Scan(&sess.ID, &sess.UserID, &sess.CSRF, &sess.ExpiresAt, &sess.CreatedAt,
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.AvatarURL, &u.Bio, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, httpx.ErrUnauthorized
		}
		return nil, nil, err
	}
	if !sess.ExpiresAt.After(clock.Now()) {
		return nil, nil, httpx.ErrUnauthorized
	}
	return &sess, &u, nil
}

func (s *Store) Revoke(ctx context.Context, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func WriteCookie(w http.ResponseWriter, sess *model.Session, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  sess.ExpiresAt,
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func ReadCookie(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
