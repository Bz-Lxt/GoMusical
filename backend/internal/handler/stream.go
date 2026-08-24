package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gomusical/internal/clock"
	"gomusical/internal/hmacx"
	"gomusical/internal/httpx"
	"gomusical/internal/stream"
)

func (a *API) OpenStream(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.TranscodeStatus != "ready" {
		httpx.Fail(w, httpx.New(409, "not_ready", "转码尚未完成"))
		return
	}
	d := a.Access.Decide(r.Context(), httpx.UserID(r.Context()), tr)
	tr.PlayCount++
	tr.UpdatedAt = clock.Now()
	_ = a.Repos.UpdateTrack(r.Context(), tr)
	ss := hmacx.NewStream(httpx.UserID(r.Context()), tr.ID, d.Tier, httpx.Fingerprint(r.Context()), d.UntilMS, a.Cfg.StreamTTL)
	tok, err := hmacx.SignStream(ss, a.Cfg.HMACSecret)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{
		"token": tok,
		"playlist": "/api/stream/" + tok + "/index.m3u8",
		"tier": d.Tier,
		"untilMs": d.UntilMS,
		"maxSegment": d.MaxSegIndex,
		"bitrate": bitrateFor(d.Tier),
		"expiresIn": int(a.Cfg.StreamTTL.Seconds()),
	})
}

func (a *API) Playlist(w http.ResponseWriter, r *http.Request) {
	ss, err := a.verifyStream(r)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	tr, err := a.Repos.TrackByID(r.Context(), ss.TrackID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	max := hmacx.MaxSegmentIndex(ss.UntilMS, stream.SegmentDurationMS)
	if tr.SegmentCount > 0 && max >= tr.SegmentCount {
		max = tr.SegmentCount - 1
	}
	body := stream.Playlist("/api/stream/"+chi.URLParam(r, "token"), tr.SegmentCount, max, stream.SegmentDurationSec)
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(body))
}

func (a *API) Segment(w http.ResponseWriter, r *http.Request) {
	ss, err := a.verifyStream(r)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	idx, ok := stream.ParseSegName(chi.URLParam(r, "seg"))
	if !ok {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	max := hmacx.MaxSegmentIndex(ss.UntilMS, stream.SegmentDurationMS)
	if idx > max {
		a.Repos.Audit(r.Context(), ss.UserID, "stream.denied", "segment_oob", map[string]any{"idx": idx, "max": max})
		httpx.Fail(w, httpx.ErrSegmentDenied)
		return
	}
	tr, err := a.Repos.TrackByID(r.Context(), ss.TrackID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	br := bitrateFor(ss.Tier)
	p, err := a.Store.Resolve(filepath.ToSlash(filepath.Join(tr.HLSDir, br, "seg_"+strconv.Itoa(idx)+".ts")))
	if err != nil {
		httpx.Fail(w, httpx.ErrNotFound)
		return
	}
	f, err := os.Open(p)
	if err != nil {
		httpx.Fail(w, httpx.ErrNotFound)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "private, max-age=60")
	http.ServeContent(w, r, "seg.ts", st.ModTime(), f)
}

func (a *API) Peaks(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.PeaksKey == "" {
		httpx.Fail(w, httpx.New(409, "not_ready", "波形尚未生成"))
		return
	}
	p, err := a.Store.Resolve(tr.PeaksKey)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	b, err := os.ReadFile(p)
	if err != nil {
		httpx.Fail(w, httpx.ErrNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(200)
	_, _ = w.Write(b)
}

func (a *API) Cover(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.CoverKey == "" || !a.Store.Exists(tr.CoverKey) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"><rect fill="#1a1410" width="200" height="200"/><circle cx="100" cy="100" r="70" fill="#2a2118" stroke="#c9a36a" stroke-width="4"/><circle cx="100" cy="100" r="12" fill="#c9a36a"/></svg>`))
		return
	}
	p, _ := a.Store.Resolve(tr.CoverKey)
	http.ServeFile(w, r, p)
}

func (a *API) verifyStream(r *http.Request) (hmacx.StreamSession, error) {
	tok := chi.URLParam(r, "token")
	if tok == "" {
		tok = strings.TrimPrefix(r.URL.Path, "/api/stream/")
		if i := strings.Index(tok, "/"); i > 0 {
			tok = tok[:i]
		}
	}
	return hmacx.VerifyStream(tok, a.Cfg.HMACSecret)
}

func bitrateFor(tier string) string {
	if tier == "PREVIEW" {
		return "128k"
	}
	return "256k"
}
