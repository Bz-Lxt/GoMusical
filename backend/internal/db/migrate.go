package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"gomusical/internal/logx"
)

const advisoryKey int64 = 0x474F4D5553 // "GOMUS"

// Migrate serializes DDL with a session-scoped advisory lock.
// sql.DB.Exec would hop connections; pgxpool.Acquire keeps lock+unlock on one conn.
func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryKey); err != nil {
		return fmt.Errorf("advisory lock: %w", err)
	}
	defer func() {
		if _, uerr := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryKey); uerr != nil {
			logx.Error("advisory unlock", "err", uerr)
		}
	}()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(raw)); err != nil {
			return fmt.Errorf("migration %s: %w", e.Name(), err)
		}
		logx.Info("migration applied", "file", e.Name())
	}
	return nil
}
