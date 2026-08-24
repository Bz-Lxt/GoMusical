package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
	"gomusical/internal/service"
	"gomusical/internal/storage"
)

type trackPatch struct {
	Title          *string `json:"title"`
	AlbumID        *string `json:"albumId"`
	PreviewSeconds *int    `json:"previewSeconds"`
	PaidDownload   *bool   `json:"paidDownload"`
	PaidPriceCents *int    `json:"paidPriceCents"`
	FanOnly        *bool   `json:"fanOnly"`
	FanDownload    *bool   `json:"fanDownload"`
	DisplayName    *string `json:"displayFilename"`
}

type albumReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

func (a *API) ListTracks(w http.ResponseWriter, r *http.Request) {
	creator := r.URL.Query().Get("creatorId")
	album := r.URL.Query().Get("albumId")
	list, err := a.Repos.ListTracks(r.Context(), creator, album)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	uid := httpx.UserID(r.Context())
	out := make([]model.Track, 0, len(list))
	for _, t := range list {
		d := a.Access.Decide(r.Context(), uid, &t)
		t.AccessTier = d.Tier
		t.AccessUntilMS = d.UntilMS
		t.StorageKey = ""
		out = append(out, t)
	}
	httpx.JSON(w, 200, map[string]any{"items": out})
}

func (a *API) GetTrack(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	d := a.Access.Decide(r.Context(), httpx.UserID(r.Context()), tr)
	tr.AccessTier = d.Tier
	tr.AccessUntilMS = d.UntilMS
	tr.StorageKey = ""
	httpx.JSON(w, 200, tr)
}

func (a *API) PatchTrack(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.CreatorID != httpx.UserID(r.Context()) && httpx.Role(r.Context()) != model.RoleAdmin {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	var req trackPatch
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	if req.Title != nil {
		tr.Title = strings.TrimSpace(*req.Title)
	}
	if req.AlbumID != nil {
		if *req.AlbumID == "" {
			tr.AlbumID = nil
		} else {
			tr.AlbumID = req.AlbumID
		}
	}
	if req.PreviewSeconds != nil {
		ps := *req.PreviewSeconds
		if ps != 15 && ps != 30 && ps != 60 {
			httpx.Fail(w, httpx.New(422, "unprocessable", "试听时长仅支持 15/30/60 秒"))
			return
		}
		tr.PreviewSeconds = ps
	}
	if req.PaidDownload != nil {
		tr.PaidDownload = *req.PaidDownload
	}
	if req.PaidPriceCents != nil {
		if *req.PaidPriceCents < 0 {
			httpx.Fail(w, httpx.ErrBadRequest)
			return
		}
		tr.PaidPriceCents = *req.PaidPriceCents
	}
	if req.FanOnly != nil {
		tr.FanOnly = *req.FanOnly
	}
	if req.FanDownload != nil {
		tr.FanDownload = *req.FanDownload
	}
	if req.DisplayName != nil {
		tr.DisplayFilename = sanitizeName(*req.DisplayName)
	}
	tr.UpdatedAt = clock.Now()
	if err := a.Repos.UpdateTrack(r.Context(), tr); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, tr)
}

func (a *API) ListAlbums(w http.ResponseWriter, r *http.Request) {
	list, err := a.Repos.ListAlbums(r.Context(), r.URL.Query().Get("creatorId"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": list})
}

func (a *API) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	var req albumReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	al := &model.Album{
		ID: uuid.NewString(), CreatorID: httpx.UserID(r.Context()),
		Title: req.Title, Description: req.Description, SortOrder: req.SortOrder,
		CreatedAt: clock.Now(),
	}
	if err := a.Repos.CreateAlbum(r.Context(), al); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 201, al)
}

func (a *API) PatchAlbum(w http.ResponseWriter, r *http.Request) {
	al, err := a.Repos.AlbumByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if al.CreatorID != httpx.UserID(r.Context()) {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	var req albumReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	if req.Title != "" {
		al.Title = req.Title
	}
	al.Description = req.Description
	al.SortOrder = req.SortOrder
	if err := a.Repos.UpdateAlbum(r.Context(), al); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, al)
}

func (a *API) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename  string `json:"filename"`
		SHA256    string `json:"sha256"`
		SizeBytes int64  `json:"sizeBytes"`
		Title     string `json:"title"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, err)
		return
	}
	var existing *model.AssetBlob
	if blob, err := a.Repos.BlobBySHA(r.Context(), req.SHA256); err == nil {
		existing = blob
	}
	plan, err := service.PlanUpload(req.SizeBytes, req.Filename, req.SHA256, existing)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if plan.Instant && existing != nil {
		tr := a.attachExisting(r, existing, req)
		httpx.JSON(w, 200, map[string]any{"instant": true, "track": tr})
		return
	}
	chunk := plan.ChunkSize
	n := plan.Chunks
	recv := make([]bool, n)
	u := &model.UploadSession{
		ID: uuid.NewString(), UserID: httpx.UserID(r.Context()),
		Filename: req.Filename, SHA256: req.SHA256, SizeBytes: req.SizeBytes,
		ChunkSize: chunk, Received: recv, TmpKey: storage.TmpKey(uuid.NewString()),
		CreatedAt: clock.Now(),
	}
	if err := a.Repos.CreateUpload(r.Context(), u); err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{
		"instant": false, "uploadId": u.ID, "chunkSize": chunk, "chunks": n, "title": req.Title,
	})
}

func (a *API) attachExisting(r *http.Request, blob *model.AssetBlob, req struct {
	Filename  string `json:"filename"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"sizeBytes"`
	Title     string `json:"title"`
}) *model.Track {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSuffix(req.Filename, filepath.Ext(req.Filename))
	}
	tr := &model.Track{
		ID: uuid.NewString(), CreatorID: httpx.UserID(r.Context()),
		Title: title, DisplayFilename: sanitizeName(req.Filename),
		ContentSHA256: blob.SHA256, StorageKey: blob.StorageKey, SizeBytes: blob.SizeBytes,
		Format: extOf(req.Filename), PreviewSeconds: 30, PaidDownload: true, PaidPriceCents: 900,
		TranscodeStatus: model.JobPending, CreatedAt: clock.Now(), UpdatedAt: clock.Now(),
	}
	_ = a.Repos.CreateTrack(r.Context(), tr)
	_, _ = a.Worker.Enqueue(r.Context(), tr.ID)
	return tr
}

func (a *API) UploadChunk(w http.ResponseWriter, r *http.Request) {
	u, err := a.Repos.UploadByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if u.UserID != httpx.UserID(r.Context()) {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	idx, _ := strconv.Atoi(r.URL.Query().Get("index"))
	if idx < 0 || idx >= len(u.Received) {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	p, err := a.Store.Resolve(u.TmpKey)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	defer f.Close()
	if _, err := f.Seek(int64(idx)*int64(u.ChunkSize), io.SeekStart); err != nil {
		httpx.Fail(w, err)
		return
	}
	n, err := io.Copy(f, io.LimitReader(r.Body, int64(u.ChunkSize)+1))
	if err != nil || n > int64(u.ChunkSize) {
		httpx.Fail(w, httpx.ErrBadRequest)
		return
	}
	_ = a.Repos.MarkChunk(r.Context(), u.ID, idx)
	httpx.JSON(w, 200, map[string]any{"index": idx, "bytes": n})
}

func (a *API) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	u, err := a.Repos.UploadByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if u.UserID != httpx.UserID(r.Context()) {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	p, err := a.Store.Resolve(u.TmpKey)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	sha, size, err := storage.HashFile(p)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if sha != u.SHA256 || size != u.SizeBytes {
		httpx.Fail(w, httpx.New(422, "hash_mismatch", "文件内容哈希与申报不一致"))
		return
	}
	ext := extOf(u.Filename)
	key := storage.BlobKey(sha, ext)
	dst, _ := a.Store.Resolve(key)
	_ = os.MkdirAll(filepath.Dir(dst), 0o755)
	if err := os.Rename(p, dst); err != nil {
		in, _ := os.Open(p)
		_, _ = a.Store.WriteFile(key, in)
		if in != nil {
			in.Close()
		}
	}
	_ = a.Repos.UpsertBlob(r.Context(), &model.AssetBlob{
		SHA256: sha, StorageKey: key, SizeBytes: size, MIME: mimeOf(ext), CreatedAt: clock.Now(),
	})
	var body struct {
		Title string `json:"title"`
	}
	_ = httpx.Decode(r, &body)
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = strings.TrimSuffix(u.Filename, filepath.Ext(u.Filename))
	}
	tr := &model.Track{
		ID: uuid.NewString(), CreatorID: u.UserID, Title: title,
		DisplayFilename: sanitizeName(u.Filename), ContentSHA256: sha, StorageKey: key,
		SizeBytes: size, Format: ext, PreviewSeconds: 30, PaidDownload: true, PaidPriceCents: 900,
		TranscodeStatus: model.JobPending, CreatedAt: clock.Now(), UpdatedAt: clock.Now(),
	}
	if err := a.Repos.CreateTrack(r.Context(), tr); err != nil {
		httpx.Fail(w, err)
		return
	}
	_, _ = a.Worker.Enqueue(r.Context(), tr.ID)
	httpx.JSON(w, 201, tr)
}

func (a *API) RetryTranscode(w http.ResponseWriter, r *http.Request) {
	tr, err := a.Repos.TrackByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	if tr.CreatorID != httpx.UserID(r.Context()) && httpx.Role(r.Context()) != model.RoleAdmin {
		httpx.Fail(w, httpx.ErrForbidden)
		return
	}
	j, err := a.Worker.Enqueue(r.Context(), tr.ID)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 202, j)
}

func (a *API) ListJobs(w http.ResponseWriter, r *http.Request) {
	uid := httpx.UserID(r.Context())
	if httpx.Role(r.Context()) == model.RoleAdmin {
		uid = ""
	}
	list, err := a.Repos.ListJobs(r.Context(), uid)
	if err != nil {
		httpx.Fail(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"items": list})
}

func sanitizeName(s string) string {
	s = filepath.Base(s)
	s = strings.ReplaceAll(s, "..", "")
	if s == "" {
		return "track.bin"
	}
	return s
}

func extOf(name string) string {
	e := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	switch e {
	case "flac", "wav", "mp3", "aac", "m4a":
		return e
	default:
		return "bin"
	}
}

func mimeOf(ext string) string {
	switch ext {
	case "flac":
		return "audio/flac"
	case "wav":
		return "audio/wav"
	case "mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
