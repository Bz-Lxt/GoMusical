package auth

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sid := ReadCookie(r)
			if sid == "" {
				next.ServeHTTP(w, r)
				return
			}
			sess, user, err := store.Get(r.Context(), sid)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, httpx.CtxUserID, user.ID)
			ctx = context.WithValue(ctx, httpx.CtxRole, user.Role)
			ctx = context.WithValue(ctx, httpx.CtxSession, sess.ID)
			ctx = context.WithValue(ctx, httpx.CtxCSRF, sess.CSRF)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpx.UserID(r.Context()) == "" {
			httpx.Fail(w, httpx.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allow := map[string]bool{}
	for _, r := range roles {
		allow[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allow[httpx.Role(r.Context())] {
				httpx.Fail(w, httpx.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if httpx.UserID(r.Context()) == "" {
			next.ServeHTTP(w, r)
			return
		}
		want, _ := r.Context().Value(httpx.CtxCSRF).(string)
		got := r.Header.Get("X-CSRF-Token")
		if want == "" || got != want {
			httpx.Fail(w, httpx.New(403, "csrf", "CSRF 校验失败"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoadUser(ctx context.Context, pool *pgxpool.Pool, id string) (*model.User, error) {
	row := pool.QueryRow(ctx, `SELECT id, email, password_hash, display_name, role, avatar_url, bio, created_at FROM users WHERE id=$1`, id)
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.AvatarURL, &u.Bio, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
