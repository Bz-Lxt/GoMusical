package handler

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"time"

	"gomusical/internal/db"
	"gomusical/internal/httpx"
	"gomusical/internal/redisx"

	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Health(pool *pgxpool.Pool, rdb *redis.Client, storeRoot, ffmpeg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		checks := map[string]string{}
		ok := true
		if err := db.Ping(ctx, pool); err != nil {
			checks["postgres"] = err.Error()
			ok = false
		} else {
			checks["postgres"] = "ok"
		}
		if err := redisx.Ping(ctx, rdb); err != nil {
			checks["redis"] = err.Error()
			ok = false
		} else {
			checks["redis"] = "ok"
		}
		if _, err := os.Stat(storeRoot); err != nil {
			checks["storage"] = err.Error()
			ok = false
		} else {
			checks["storage"] = "ok"
		}
		if _, err := exec.LookPath(ffmpeg); err != nil {
			if _, err2 := os.Stat(ffmpeg); err2 != nil {
				checks["ffmpeg"] = "missing"
				ok = false
			} else {
				checks["ffmpeg"] = "ok"
			}
		} else {
			checks["ffmpeg"] = "ok"
		}
		status := 200
		if !ok {
			status = 503
		}
		httpx.JSON(w, status, map[string]any{"ok": ok, "checks": checks})
	}
}
