package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
	"gomusical/internal/stats"
)

func (a *API) ListComments(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	hidden := httpx.UserID(r.Context()) == tr.CreatorID || httpx.Role(r.Context()) == model.RoleAdmin
	list, err := a.Repos.ListComments(r.Context(), tr.ID, hidden)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": list})
}

func (a *API) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TimestampMS int    `json:"timestampMs"`
		Body        string `json:"body"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if !httpx.CommentBody(req.Body) {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if err := a.Access.AssertCommentWindow(r.Context(), httpx.UserID(r.Context()), tr, req.TimestampMS); err != nil {
		httpx.Fail(w, err)
		return
	}
	c := &model.Comment{
		ID: uuid.NewString(), TrackID: tr.ID, UserID: httpx.UserID(r.Context()),
		TimestampMS: req.TimestampMS, Body: req.Body, CreatedAt: clock.Now(),
	}
	if err := a.Repos.CreateComment(r.Context(), c); err != nil {
		httpx.Fail(w, err)
		return
	}
	u, _ := a.Repos.UserByID(r.Context(), c.UserID)
	if u != nil {
		c.AuthorName = u.DisplayName
	}
	httpx.JSON(w, 201, c)
}

func (a *API) ModComment(w http.ResponseWriter, r *http.Request) {
	c, err := a.Repos.CommentByID(r.Context(), chi.URLParam(r, "cid"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	tr, err := a.Repos.TrackByID(r.Context(), c.TrackID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.CreatorID != httpx.UserID(r.Context()) && httpx.Role(r.Context()) != model.RoleAdmin {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	var req struct {
		Pinned *bool   `json:"pinned"`
		Hidden *bool   `json:"hidden"`
		Reply  *string `json:"reply"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	if req.Pinned != nil {
		c.Pinned = *req.Pinned
	}
	if req.Hidden != nil {
		c.Hidden = *req.Hidden
	}
	if req.Reply != nil {
		c.Reply = *req.Reply
	}
	if err := a.Repos.UpdateCommentMod(r.Context(), c.ID, c.Pinned, c.Hidden, c.Reply); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, c)
}

func (a *API) LikeComment(w http.ResponseWriter, r *http.Request) {
	if err := a.Repos.LikeComment(r.Context(), chi.URLParam(r, "cid")); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"ok": true})
}

func (a *API) SponsorTrack(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AmountCents int `json:"amountCents"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	out, err := a.Sponsor.CreateTrackSponsor(r.Context(), httpx.UserID(r.Context()), chi.URLParam(r, "id"), req.AmountCents)
	if err != nil && out == nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, out)
}

func (a *API) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AmountCents int `json:"amountCents"`
	}
	_ = httpx.Decode(r, &req)
	out, err := a.Sponsor.CreateFanSub(r.Context(), httpx.UserID(r.Context()), chi.URLParam(r, "id"), req.AmountCents)
	if err != nil && out == nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, out)
}

func (a *API) PayCallback(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := httpx.Decode(r, &payload); err != nil {
		httpx.Fail(w, err)
		return
	}
	if err := a.Sponsor.Callback(r.Context(), payload); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"ok": true})
}

func (a *API) MyOrders(w http.ResponseWriter, r *http.Request) {
	list, err := a.Repos.ListOrdersByUser(r.Context(), httpx.UserID(r.Context()))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": list})
}

func (a *API) CreatorOrders(w http.ResponseWriter, r *http.Request) {
	list, err := a.Repos.ListOrdersByCreator(r.Context(), httpx.UserID(r.Context()))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": list})
}

func (a *API) ListCreators(w http.ResponseWriter, r *http.Request) {
	list, err := a.Repos.ListCreators(r.Context())
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	pub := make([]map[string]any, 0, len(list))
	for i := range list {
		pub = append(pub, publicUser(&list[i]))
	}
	httpx.JSON(w, 200, map[string]any{"items": pub})
}

func (a *API) GetCreator(w http.ResponseWriter, r *http.Request) {
	u, err := a.Repos.UserByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	tracks, _ := a.Repos.ListTracks(r.Context(), u.ID, "")
	httpx.JSON(w, 200, map[string]any{"creator": publicUser(u), "tracks": tracks})
}

func (a *API) AdminStats(w http.ResponseWriter, r *http.Request) {
	st, err := stats.Collect(r.Context(), a.Repos.Pool)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	logs, _ := a.Repos.ListAudit(r.Context(), 80)
	httpx.JSON(w, 200, map[string]any{"stats": st, "audit": logs})
}
