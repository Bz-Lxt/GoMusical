package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gomusical/internal/auth"
	"gomusical/internal/config"
	"gomusical/internal/db"
	"gomusical/internal/download"
	"gomusical/internal/handler"
	"gomusical/internal/logx"
	"gomusical/internal/payment"
	"gomusical/internal/redisx"
	"gomusical/internal/repo"
	"gomusical/internal/seed"
	"gomusical/internal/service"
	"gomusical/internal/storage"
	"gomusical/internal/transcode"
)

func main() {
	cfg := config.Load()
	logx.Init(cfg.LogLevel, cfg.Env)
	if err := cfg.Validate(); err != nil {
		logx.Error("invalid config", "err", err)
		os.Exit(1)
	}
	if cfg.PaymentMode == "mock" {
		logx.Warn("payment running in mock mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logx.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	migDir := os.Getenv("MIGRATIONS_DIR")
	if migDir == "" {
		migDir = "/app/migrations"
	}
	if err := db.Migrate(ctx, pool, migDir); err != nil {
		logx.Error("migrate", "err", err)
		os.Exit(1)
	}

	rdb := redisx.Connect(cfg.RedisAddr, cfg.RedisPass)
	store, err := storage.New(cfg.StorageRoot)
	if err != nil {
		logx.Error("storage", "err", err)
		os.Exit(1)
	}
	repos := &repo.Repos{Pool: pool}
	sess := &auth.Store{Pool: pool, TTL: cfg.SessionTTL}
	eng := &transcode.Engine{FFmpeg: cfg.FFmpegBin, FFprobe: cfg.FFprobeBin, Store: store}
	worker := &transcode.Worker{Eng: eng, Repos: repos, Store: store}
	worker.Start(context.Background())
	defer worker.Stop()

	pay := payment.Select(cfg.PaymentMode, cfg.PaymentRealKey, cfg.PublicOrigin, cfg.MockPayBehavior)
	access := &service.Access{Repos: repos}
	sp := &service.Sponsor{Repos: repos, Pay: pay, Cfg: cfg, Access: access}
	lim := download.NewLimiter(rdb, cfg.DownloadConc, cfg.DailyDownloads, cfg.UserBPS, cfg.GlobalBPS)

	if cfg.SeedEnabled {
		if err := seed.Run(context.Background(), repos, store, eng, worker, cfg.FFmpegBin); err != nil {
			logx.Error("seed", "err", err)
		}
	}

	api := &handler.API{
		Cfg: cfg, Repos: repos, Sess: sess, Store: store,
		Access: access, Sponsor: sp, Worker: worker, Limiter: lim, Eng: eng,
	}
	h := handler.NewRouter(api, pool, rdb)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 8 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		logx.Info("listening", "addr", cfg.HTTPAddr, "origin", cfg.PublicOrigin)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	shctx, shcancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer shcancel()
	_ = srv.Shutdown(shctx)
}
