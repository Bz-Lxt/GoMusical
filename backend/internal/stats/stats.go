package stats

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
	"gomusical/internal/clock"
)

type Snapshot struct {
	Users          int     `json:"users"`
	Tracks         int     `json:"tracks"`
	ReadyTracks    int     `json:"readyTracks"`
	PaidOrders     int     `json:"paidOrders"`
	SponsorCents   int64   `json:"sponsorCents"`
	PlayCount      int64   `json:"playCount"`
	Comments       int     `json:"comments"`
	ConversionRate float64 `json:"previewToSponsorRate"`
	AuditDenies    int     `json:"auditDenies"`
	GeneratedAt    string  `json:"generatedAt"`
}

func Collect(ctx context.Context, pool *pgxpool.Pool) (Snapshot, error) {
	var s Snapshot
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.Users)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&s.Tracks)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM tracks WHERE transcode_status='ready'`).Scan(&s.ReadyTracks)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE status='paid'`).Scan(&s.PaidOrders)
	_ = pool.QueryRow(ctx, `SELECT COALESCE(SUM(sponsor_cents),0) FROM tracks`).Scan(&s.SponsorCents)
	_ = pool.QueryRow(ctx, `SELECT COALESCE(SUM(play_count),0) FROM tracks`).Scan(&s.PlayCount)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM comments`).Scan(&s.Comments)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE action IN ('stream.denied','ticket.verify_fail')`).Scan(&s.AuditDenies)
	if s.PlayCount > 0 {
		s.ConversionRate = math.Round(float64(s.PaidOrders)/float64(s.PlayCount)*10000) / 10000
	}
	s.GeneratedAt = clock.FormatDisplay(clock.Now())
	return s, nil
}

func CreatorBoard(ctx context.Context, pool *pgxpool.Pool, creatorID string) (map[string]any, error) {
	var plays, cents int64
	var tracks, comments int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*), COALESCE(SUM(play_count),0), COALESCE(SUM(sponsor_cents),0) FROM tracks WHERE creator_id=$1`, creatorID).
		Scan(&tracks, &plays, &cents)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM comments c JOIN tracks t ON t.id=c.track_id WHERE t.creator_id=$1 AND c.hidden=false`, creatorID).
		Scan(&comments)
	return map[string]any{
		"tracks": tracks, "plays": plays, "sponsorCents": cents, "openComments": comments,
		"generatedAt": clock.FormatDisplay(clock.Now()),
	}, nil
}
