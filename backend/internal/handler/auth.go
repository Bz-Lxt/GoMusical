package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gomusical/internal/auth"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if !httpx.Email(req.Email) || len(req.Password) < 8 || !httpx.DisplayName(req.DisplayName) {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role != model.RoleCreator && role != model.RoleListener {
		role = model.RoleListener
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.Fail(w, httpx.Wrap(400, "bad_request", "密码不符合要求", err))
		return
	}
	u := &model.User{
		ID: uuid.NewString(), Email: req.Email, PasswordHash: hash,
		DisplayName: req.DisplayName, Role: role, CreatedAt: clock.Now(),
	}
	if err := a.Repos.CreateUser(r.Context(), u); err != nil {
		httpx.Fail(w, err)
		return
	}
	sess, err := a.Sess.Issue(r.Context(), u.ID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	auth.WriteCookie(w, sess, false)
	httpx.JSON(w, 201, map[string]any{"user": publicUser(u), "csrf": sess.CSRF})
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	u, err := a.Repos.UserByEmail(r.Context(), req.Email)
	if err != nil || !auth.VerifyPassword(req.Password, u.PasswordHash) {
		httpx.Fail(w, httpx.New(401, "unauthorized", "邮箱或密码错误"))
		return
	}
	sess, err := a.Sess.Issue(r.Context(), u.ID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	auth.WriteCookie(w, sess, false)
	httpx.JSON(w, 200, map[string]any{"user": publicUser(u), "csrf": sess.CSRF})
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	if sid := auth.ReadCookie(r); sid != "" {
		_ = a.Sess.Revoke(r.Context(), sid)
	}
	auth.ClearCookie(w)
	httpx.JSON(w, 200, map[string]any{"ok": true})
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	uid := httpx.UserID(r.Context())
	if uid == "" {
		httpx.Fail(w, httpx.ErrUnauthorized)
		return
	}
	u, err := a.Repos.UserByID(r.Context(), uid)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	csrf, _ := r.Context().Value(httpx.CtxCSRF).(string)
	httpx.JSON(w, 200, map[string]any{"user": publicUser(u), "csrf": csrf})
}

func publicUser(u *model.User) map[string]any {
	return map[string]any{
		"id": u.ID, "email": u.Email, "displayName": u.DisplayName,
		"role": u.Role, "avatarUrl": u.AvatarURL, "bio": u.Bio,
		"createdAt": clock.FormatDisplay(u.CreatedAt),
	}
}
