package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gomusical/internal/auth"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

func NewRouter(a *API, pool *pgxpool.Pool, rdb *redis.Client) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.RequestID)
	r.Use(httpx.Recover)
	r.Use(httpx.AccessLog)
	r.Use(httpx.CORS(a.Cfg.AllowedOrigins))
	r.Use(httpx.ClientFingerprint)
	r.Use(auth.Middleware(a.Sess))
	r.Use(auth.CSRF)
	r.Use(httpx.RefererGuard(a.Cfg.RefererAllow, []string{"/api/stream/"}))

	r.Get("/api/health", Health(pool, rdb, a.Cfg.StorageRoot, a.Cfg.FFmpegBin))

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", a.Register)
		r.Post("/login", a.Login)
		r.Post("/logout", a.Logout)
		r.Get("/me", a.Me)
	})

	r.Get("/api/creators", a.ListCreators)
	r.Get("/api/creators/{id}", a.GetCreator)
	r.Get("/api/tracks", a.ListTracks)
	r.Get("/api/tracks/{id}", a.GetTrack)
	r.Get("/api/tracks/{id}/peaks", a.Peaks)
	r.Get("/api/tracks/{id}/cover", a.Cover)
	r.Get("/api/tracks/{id}/comments", a.ListComments)
	r.Get("/api/albums", a.ListAlbums)
	r.Get("/api/stream/{id}/open", a.OpenStream)
	r.Get("/api/stream/{token}/index.m3u8", a.Playlist)
	r.Get("/api/stream/{token}/{seg}", a.Segment)
	r.Get("/api/download/*", a.Download)

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Post("/api/tracks/{id}/comments", a.CreateComment)
		r.Post("/api/comments/{cid}/like", a.LikeComment)
		r.Post("/api/tracks/{id}/sponsor", a.SponsorTrack)
		r.Post("/api/creators/{id}/subscribe", a.Subscribe)
		r.Post("/api/pay/callback", a.PayCallback)
		r.Get("/api/me/orders", a.MyOrders)
		r.Post("/api/tracks/{id}/ticket", a.IssueTicket)
		r.Get("/api/tickets/{nonce}", a.TicketStatus)
		r.Post("/api/tickets/{nonce}/revoke", a.RevokeTicket)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(auth.RequireRole(model.RoleCreator, model.RoleAdmin))
		r.Post("/api/albums", a.CreateAlbum)
		r.Patch("/api/albums/{id}", a.PatchAlbum)
		r.Patch("/api/tracks/{id}", a.PatchTrack)
		r.Post("/api/uploads", a.InitUpload)
		r.Put("/api/uploads/{id}/chunk", a.UploadChunk)
		r.Post("/api/uploads/{id}/complete", a.CompleteUpload)
		r.Post("/api/tracks/{id}/transcode", a.RetryTranscode)
		r.Get("/api/jobs", a.ListJobs)
		r.Get("/api/creator/orders", a.CreatorOrders)
		r.Patch("/api/comments/{cid}", a.ModComment)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(auth.RequireRole(model.RoleAdmin))
		r.Get("/api/admin/stats", a.AdminStats)
	})

	return r
}
