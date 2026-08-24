package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"gomusical/internal/clock"
	"gomusical/internal/download"
	"gomusical/internal/hmacx"
	"gomusical/internal/httpx"
)

func (a *API) IssueTicket(w http.ResponseWriter, r *http.Request) {
	raw, tk, err := a.Sponsor.IssueTicket(r.Context(), httpx.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{
		"ticket": raw,
		"url": "/api/download/" + raw,
		"expiresAt": clock.FormatDisplay(clock.Now().Add(a.Cfg.TicketTTL)),
		"maxUses": tk.MaxUses,
		"nonce": tk.Nonce,
		"ttlSec": int(a.Cfg.TicketTTL.Seconds()),
	})
}

func (a *API) Download(w http.ResponseWriter, r *http.Request) {
	raw := chi.URLParam(r, "*")
	raw = strings.TrimPrefix(raw, "/")
	tk, err := hmacx.VerifyTicket(raw, a.Cfg.HMACSecret)
	if err != nil {
		a.Repos.Audit(r.Context(), "", "ticket.verify_fail", err.Error(), map[string]any{"ip": r.RemoteAddr})
		a.Limiter.FlagAbuse(r.Context(), r.RemoteAddr, "tamper")
		httpx.Fail(w, err)
		return
	}
	row, err := a.Repos.TicketByNonce(r.Context(), tk.Nonce)
	if err != nil {
		httpx.Fail(w, httpx.ErrTicketTampered)
		return
	}
	if row.Revoked {
		httpx.Fail(w, httpx.ErrTicketRevoked)
		return
	}
	policy := download.DefaultResume()
	if row.Uses >= row.MaxUses {
		trChk, _ := a.Repos.TrackByID(r.Context(), tk.TrackID)
		sz := int64(0)
		if trChk != nil {
			sz = trChk.SizeBytes
		}
		if !policy.AllowAnotherHit(row.Uses, row.MaxUses, row.BytesDone, sz) {
			httpx.Fail(w, httpx.ErrTicketExhausted)
			return
		}
	}
	tr, err := a.Repos.TrackByID(r.Context(), tk.TrackID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tk.UserID != tr.CreatorID && tk.TrackID != tr.ID {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	release, err := a.Limiter.Acquire(r.Context(), tk.UserID)
	if err != nil {
		w.Header().Set("Retry-After", "3")
		httpx.Fail(w, err)
		return
	}
	defer release()

	p, err := a.Store.Resolve(tr.StorageKey)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	f, err := os.Open(p)
	if err != nil {
		httpx.Fail(w, httpx.ErrNotFound)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		httpx.Fail(w, err)
		return
	}

	var written atomic.Int64
	mw := &download.Metered{W: w, OnByte: func(n int) {
		written.Add(int64(n))
		a.Limiter.WaitUser(tk.UserID, n)
		a.Limiter.WaitBytes(n)
	}}
	rl := &download.RateLimited{Src: f, Wait: nil}
	name := tr.DisplayFilename
	if name == "" {
		name = "track." + tr.Format
	}
	mw.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(name, `"`, "")+`"`)
	mw.Header().Set("X-Content-SHA256", tr.ContentSHA256)
	mw.Header().Set("Accept-Ranges", "bytes")
	// ServeContent uses the ReadSeeker; metering wraps ResponseWriter.
	http.ServeContent(mw, r, filepath.Base(name), st.ModTime(), rl)

	n := written.Load()
	if n == 0 {
		n = download.BytesOfRange(r.Header.Get("Range"), st.Size())
	}
	total := row.BytesDone + n
	complete := total >= int64(float64(st.Size())*0.95)
	already := row.BytesDone >= int64(float64(st.Size())*0.95)
	if err := a.Repos.AddTicketBytes(r.Context(), tk.Nonce, n, complete && !already); err != nil {
		return
	}
	if complete && !already {
		a.Limiter.MarkComplete(r.Context(), tk.UserID)
		a.Repos.Audit(r.Context(), tk.UserID, "download.complete", "ok", map[string]any{"track": tr.ID, "bytes": total})
	}
}

func (a *API) RevokeTicket(w http.ResponseWriter, r *http.Request) {
	nonce := chi.URLParam(r, "nonce")
	row, err := a.Repos.TicketByNonce(r.Context(), nonce)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if row.UserID != httpx.UserID(r.Context()) && httpx.Role(r.Context()) != "ADMIN" && httpx.Role(r.Context()) != "CREATOR" {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	if err := a.Repos.RevokeTicket(r.Context(), nonce); err != nil {
		httpx.Fail(w, err)
		return
	}
	a.Repos.Audit(r.Context(), httpx.UserID(r.Context()), "ticket.revoke", "manual", map[string]any{"nonce": nonce})
	httpx.JSON(w, 200, map[string]any{"revoked": true})
}

func (a *API) TicketStatus(w http.ResponseWriter, r *http.Request) {
	row, err := a.Repos.TicketByNonce(r.Context(), chi.URLParam(r, "nonce"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{
		"nonce": row.Nonce, "uses": row.Uses, "maxUses": row.MaxUses,
		"bytesDone": row.BytesDone, "revoked": row.Revoked,
		"expiresAt": clock.FormatDisplay(row.ExpiresAt),
		"concurrency": a.Limiter.Concurrent(r.Context(), row.UserID),
		"dailyUsed": a.Limiter.DailyUsed(r.Context(), row.UserID),
	})
}
