package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

type Repos struct {
	Pool *pgxpool.Pool
}

func (r *Repos) CreateUser(ctx context.Context, u *model.User) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, role, avatar_url, bio, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		u.ID, strings.ToLower(u.Email), u.PasswordHash, u.DisplayName, u.Role, u.AvatarURL, u.Bio, u.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return httpx.New(409, "conflict", "邮箱已注册")
	}
	return err
}

func (r *Repos) UserByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, role, avatar_url, bio, created_at
		FROM users WHERE email=$1`, strings.ToLower(email))
	return scanUser(row)
}

func (r *Repos) UserByID(ctx context.Context, id string) (*model.User, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, role, avatar_url, bio, created_at
		FROM users WHERE id=$1`, id)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.AvatarURL, &u.Bio, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repos) ListCreators(ctx context.Context) ([]model.User, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, email, password_hash, display_name, role, avatar_url, bio, created_at
		FROM users WHERE role='CREATOR' ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.AvatarURL, &u.Bio, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repos) CreateAlbum(ctx context.Context, a *model.Album) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO albums (id, creator_id, title, cover_key, description, sort_order, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.CreatorID, a.Title, a.CoverKey, a.Description, a.SortOrder, a.CreatedAt)
	return err
}

func (r *Repos) UpdateAlbum(ctx context.Context, a *model.Album) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE albums SET title=$2, description=$3, sort_order=$4, cover_key=$5 WHERE id=$1 AND creator_id=$6`,
		a.ID, a.Title, a.Description, a.SortOrder, a.CoverKey, a.CreatorID)
	return err
}

func (r *Repos) AlbumByID(ctx context.Context, id string) (*model.Album, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT a.id, a.creator_id, a.title, a.cover_key, a.description, a.sort_order, a.created_at,
		       (SELECT COUNT(*) FROM tracks t WHERE t.album_id=a.id)
		FROM albums a WHERE a.id=$1`, id)
	var a model.Album
	if err := row.Scan(&a.ID, &a.CreatorID, &a.Title, &a.CoverKey, &a.Description, &a.SortOrder, &a.CreatedAt, &a.TrackCount); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *Repos) ListAlbums(ctx context.Context, creatorID string) ([]model.Album, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT a.id, a.creator_id, a.title, a.cover_key, a.description, a.sort_order, a.created_at,
		       (SELECT COUNT(*) FROM tracks t WHERE t.album_id=a.id)
		FROM albums a WHERE ($1='' OR a.creator_id=$1) ORDER BY a.sort_order, a.created_at DESC`, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Album{}
	for rows.Next() {
		var a model.Album
		if err := rows.Scan(&a.ID, &a.CreatorID, &a.Title, &a.CoverKey, &a.Description, &a.SortOrder, &a.CreatedAt, &a.TrackCount); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repos) CreateTrack(ctx context.Context, t *model.Track) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO tracks (
			id, creator_id, album_id, title, display_filename, duration_ms, format,
			content_sha256, storage_key, size_bytes, preview_seconds, paid_download,
			paid_price_cents, fan_only, fan_download, play_count, sponsor_cents,
			transcode_status, transcode_error, peaks_key, hls_dir, cover_key, segment_count,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25
		)`,
		t.ID, t.CreatorID, t.AlbumID, t.Title, t.DisplayFilename, t.DurationMS, t.Format,
		t.ContentSHA256, t.StorageKey, t.SizeBytes, t.PreviewSeconds, t.PaidDownload,
		t.PaidPriceCents, t.FanOnly, t.FanDownload, t.PlayCount, t.SponsorCents,
		t.TranscodeStatus, t.TranscodeError, t.PeaksKey, t.HLSDir, t.CoverKey, t.SegmentCount,
		t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Repos) UpdateTrack(ctx context.Context, t *model.Track) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE tracks SET
			album_id=$2, title=$3, display_filename=$4, preview_seconds=$5,
			paid_download=$6, paid_price_cents=$7, fan_only=$8, fan_download=$9,
			duration_ms=$10, format=$11, content_sha256=$12, storage_key=$13, size_bytes=$14,
			transcode_status=$15, transcode_error=$16, peaks_key=$17, hls_dir=$18,
			cover_key=$19, segment_count=$20, play_count=$21, sponsor_cents=$22, updated_at=$23
		WHERE id=$1`,
		t.ID, t.AlbumID, t.Title, t.DisplayFilename, t.PreviewSeconds,
		t.PaidDownload, t.PaidPriceCents, t.FanOnly, t.FanDownload,
		t.DurationMS, t.Format, t.ContentSHA256, t.StorageKey, t.SizeBytes,
		t.TranscodeStatus, t.TranscodeError, t.PeaksKey, t.HLSDir,
		t.CoverKey, t.SegmentCount, t.PlayCount, t.SponsorCents, t.UpdatedAt)
	return err
}

func trackSelect() string {
	return `t.id, t.creator_id, t.album_id, t.title, t.display_filename, t.duration_ms, t.format,
		t.content_sha256, t.storage_key, t.size_bytes, t.preview_seconds, t.paid_download,
		t.paid_price_cents, t.fan_only, t.fan_download, t.play_count, t.sponsor_cents,
		t.transcode_status, t.transcode_error, t.peaks_key, t.hls_dir, t.cover_key, t.segment_count,
		t.created_at, t.updated_at, COALESCE(u.display_name,'')`
}

func scanTrack(row pgx.Row) (*model.Track, error) {
	var t model.Track
	if err := row.Scan(
		&t.ID, &t.CreatorID, &t.AlbumID, &t.Title, &t.DisplayFilename, &t.DurationMS, &t.Format,
		&t.ContentSHA256, &t.StorageKey, &t.SizeBytes, &t.PreviewSeconds, &t.PaidDownload,
		&t.PaidPriceCents, &t.FanOnly, &t.FanDownload, &t.PlayCount, &t.SponsorCents,
		&t.TranscodeStatus, &t.TranscodeError, &t.PeaksKey, &t.HLSDir, &t.CoverKey, &t.SegmentCount,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatorName,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *Repos) TrackByID(ctx context.Context, id string) (*model.Track, error) {
	row := r.Pool.QueryRow(ctx, `SELECT `+trackSelect()+` FROM tracks t JOIN users u ON u.id=t.creator_id WHERE t.id=$1`, id)
	return scanTrack(row)
}

func (r *Repos) ListTracks(ctx context.Context, creatorID, albumID string) ([]model.Track, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT `+trackSelect()+`
		FROM tracks t JOIN users u ON u.id=t.creator_id
		WHERE ($1='' OR t.creator_id=$1) AND ($2='' OR t.album_id=$2)
		ORDER BY t.created_at DESC`, creatorID, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Track{}
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *Repos) BlobBySHA(ctx context.Context, sha string) (*model.AssetBlob, error) {
	row := r.Pool.QueryRow(ctx, `SELECT sha256, storage_key, size_bytes, mime, created_at FROM asset_blobs WHERE sha256=$1`, sha)
	var b model.AssetBlob
	if err := row.Scan(&b.SHA256, &b.StorageKey, &b.SizeBytes, &b.MIME, &b.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *Repos) UpsertBlob(ctx context.Context, b *model.AssetBlob) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO asset_blobs (sha256, storage_key, size_bytes, mime, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (sha256) DO UPDATE SET storage_key=EXCLUDED.storage_key`,
		b.SHA256, b.StorageKey, b.SizeBytes, b.MIME, b.CreatedAt)
	return err
}

func (r *Repos) CreateComment(ctx context.Context, c *model.Comment) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO comments (id, track_id, user_id, timestamp_ms, body, likes, pinned, hidden, reply, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.TrackID, c.UserID, c.TimestampMS, c.Body, c.Likes, c.Pinned, c.Hidden, c.Reply, c.CreatedAt)
	return err
}

func (r *Repos) ListComments(ctx context.Context, trackID string, includeHidden bool) ([]model.Comment, error) {
	q := `
		SELECT c.id, c.track_id, c.user_id, COALESCE(u.display_name,''), c.timestamp_ms, c.body, c.likes, c.pinned, c.hidden, c.reply, c.created_at
		FROM comments c JOIN users u ON u.id=c.user_id
		WHERE c.track_id=$1`
	if !includeHidden {
		q += ` AND c.hidden=false`
	}
	q += ` ORDER BY c.pinned DESC, c.timestamp_ms ASC`
	rows, err := r.Pool.Query(ctx, q, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Comment{}
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.TrackID, &c.UserID, &c.AuthorName, &c.TimestampMS, &c.Body, &c.Likes, &c.Pinned, &c.Hidden, &c.Reply, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repos) CommentByID(ctx context.Context, id string) (*model.Comment, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT c.id, c.track_id, c.user_id, COALESCE(u.display_name,''), c.timestamp_ms, c.body, c.likes, c.pinned, c.hidden, c.reply, c.created_at
		FROM comments c JOIN users u ON u.id=c.user_id WHERE c.id=$1`, id)
	var c model.Comment
	if err := row.Scan(&c.ID, &c.TrackID, &c.UserID, &c.AuthorName, &c.TimestampMS, &c.Body, &c.Likes, &c.Pinned, &c.Hidden, &c.Reply, &c.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *Repos) UpdateCommentMod(ctx context.Context, id string, pinned, hidden bool, reply string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE comments SET pinned=$2, hidden=$3, reply=$4 WHERE id=$1`, id, pinned, hidden, reply)
	return err
}

func (r *Repos) LikeComment(ctx context.Context, id string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE comments SET likes=likes+1 WHERE id=$1`, id)
	return err
}

func (r *Repos) CreateOrder(ctx context.Context, o *model.Order) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO orders (id, order_no, user_id, track_id, creator_id, kind, amount_cents, status, provider, created_at, paid_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		o.ID, o.OrderNo, o.UserID, o.TrackID, o.CreatorID, o.Kind, o.AmountCents, o.Status, o.Provider, o.CreatedAt, o.PaidAt)
	return err
}

func (r *Repos) OrderByNo(ctx context.Context, no string) (*model.Order, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, order_no, user_id, track_id, creator_id, kind, amount_cents, status, provider, created_at, paid_at
		FROM orders WHERE order_no=$1`, no)
	return scanOrder(row)
}

func scanOrder(row pgx.Row) (*model.Order, error) {
	var o model.Order
	if err := row.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.TrackID, &o.CreatorID, &o.Kind, &o.AmountCents, &o.Status, &o.Provider, &o.CreatedAt, &o.PaidAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *Repos) MarkOrderPaid(ctx context.Context, orderNo string, paidAt any) (bool, error) {
	tag, err := r.Pool.Exec(ctx, `UPDATE orders SET status='paid', paid_at=$2 WHERE order_no=$1 AND status<>'paid'`, orderNo, paidAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *Repos) ListOrdersByUser(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, order_no, user_id, track_id, creator_id, kind, amount_cents, status, provider, created_at, paid_at
		FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

func (r *Repos) ListOrdersByCreator(ctx context.Context, creatorID string) ([]model.Order, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT o.id, o.order_no, o.user_id, o.track_id, o.creator_id, o.kind, o.amount_cents, o.status, o.provider, o.created_at, o.paid_at
		FROM orders o
		LEFT JOIN tracks t ON t.id=o.track_id
		WHERE o.creator_id=$1 OR t.creator_id=$1
		ORDER BY o.created_at DESC`, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

func (r *Repos) CreateGrant(ctx context.Context, g *model.Grant) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO grants (id, user_id, track_id, creator_id, kind, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		g.ID, g.UserID, g.TrackID, g.CreatorID, g.Kind, g.ExpiresAt, g.CreatedAt)
	return err
}

func (r *Repos) GrantsFor(ctx context.Context, userID, trackID string) ([]model.Grant, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, user_id, track_id, creator_id, kind, expires_at, created_at
		FROM grants WHERE user_id=$1 AND (track_id=$2 OR track_id IS NULL)`, userID, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Grant{}
	for rows.Next() {
		var g model.Grant
		if err := rows.Scan(&g.ID, &g.UserID, &g.TrackID, &g.CreatorID, &g.Kind, &g.ExpiresAt, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *Repos) UpsertSub(ctx context.Context, s *model.Subscription) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, creator_id, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (user_id, creator_id) DO UPDATE SET expires_at=EXCLUDED.expires_at`,
		s.ID, s.UserID, s.CreatorID, s.ExpiresAt, s.CreatedAt)
	return err
}

func (r *Repos) ActiveSub(ctx context.Context, userID, creatorID string) (*model.Subscription, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, user_id, creator_id, expires_at, created_at
		FROM subscriptions WHERE user_id=$1 AND creator_id=$2 AND expires_at > $3`,
		userID, creatorID, clock.Now())
	var s model.Subscription
	if err := row.Scan(&s.ID, &s.UserID, &s.CreatorID, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repos) InsertTicket(ctx context.Context, t *model.DownloadTicketRow) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO download_tickets (nonce, grant_id, user_id, track_id, max_uses, uses, bytes_done, revoked, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.Nonce, t.GrantID, t.UserID, t.TrackID, t.MaxUses, t.Uses, t.BytesDone, t.Revoked, t.ExpiresAt, t.CreatedAt)
	return err
}

func (r *Repos) TicketByNonce(ctx context.Context, nonce string) (*model.DownloadTicketRow, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT nonce, grant_id, user_id, track_id, max_uses, uses, bytes_done, revoked, expires_at, created_at
		FROM download_tickets WHERE nonce=$1`, nonce)
	var t model.DownloadTicketRow
	if err := row.Scan(&t.Nonce, &t.GrantID, &t.UserID, &t.TrackID, &t.MaxUses, &t.Uses, &t.BytesDone, &t.Revoked, &t.ExpiresAt, &t.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *Repos) AddTicketBytes(ctx context.Context, nonce string, add int64, complete bool) error {
	if complete {
		_, err := r.Pool.Exec(ctx, `UPDATE download_tickets SET bytes_done=bytes_done+$2, uses=uses+1 WHERE nonce=$1`, nonce, add)
		return err
	}
	_, err := r.Pool.Exec(ctx, `UPDATE download_tickets SET bytes_done=bytes_done+$2 WHERE nonce=$1`, nonce, add)
	return err
}

func (r *Repos) RevokeTicket(ctx context.Context, nonce string) error {
	_, err := r.Pool.Exec(ctx, `UPDATE download_tickets SET revoked=true WHERE nonce=$1`, nonce)
	return err
}

func (r *Repos) ListTicketsByGrant(ctx context.Context, grantID string) ([]model.DownloadTicketRow, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT nonce, grant_id, user_id, track_id, max_uses, uses, bytes_done, revoked, expires_at, created_at
		FROM download_tickets WHERE grant_id=$1 ORDER BY created_at DESC`, grantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.DownloadTicketRow{}
	for rows.Next() {
		var t model.DownloadTicketRow
		if err := rows.Scan(&t.Nonce, &t.GrantID, &t.UserID, &t.TrackID, &t.MaxUses, &t.Uses, &t.BytesDone, &t.Revoked, &t.ExpiresAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repos) CreateJob(ctx context.Context, j *model.TranscodeJob) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO transcode_jobs (id, track_id, status, progress, error, attempts, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		j.ID, j.TrackID, j.Status, j.Progress, j.Error, j.Attempts, j.CreatedAt, j.UpdatedAt)
	return err
}

func (r *Repos) UpdateJob(ctx context.Context, j *model.TranscodeJob) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE transcode_jobs SET status=$2, progress=$3, error=$4, attempts=$5, updated_at=$6 WHERE id=$1`,
		j.ID, j.Status, j.Progress, j.Error, j.Attempts, j.UpdatedAt)
	return err
}

func (r *Repos) LatestJob(ctx context.Context, trackID string) (*model.TranscodeJob, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, track_id, status, progress, error, attempts, created_at, updated_at
		FROM transcode_jobs WHERE track_id=$1 ORDER BY created_at DESC LIMIT 1`, trackID)
	var j model.TranscodeJob
	if err := row.Scan(&j.ID, &j.TrackID, &j.Status, &j.Progress, &j.Error, &j.Attempts, &j.CreatedAt, &j.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &j, nil
}

func (r *Repos) ListJobs(ctx context.Context, creatorID string) ([]model.TranscodeJob, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT j.id, j.track_id, j.status, j.progress, j.error, j.attempts, j.created_at, j.updated_at
		FROM transcode_jobs j JOIN tracks t ON t.id=j.track_id
		WHERE ($1='' OR t.creator_id=$1)
		ORDER BY j.created_at DESC LIMIT 100`, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TranscodeJob{}
	for rows.Next() {
		var j model.TranscodeJob
		if err := rows.Scan(&j.ID, &j.TrackID, &j.Status, &j.Progress, &j.Error, &j.Attempts, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *Repos) PendingJobs(ctx context.Context) ([]model.TranscodeJob, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, track_id, status, progress, error, attempts, created_at, updated_at
		FROM transcode_jobs WHERE status IN ('pending','failed') AND attempts < 3
		ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TranscodeJob{}
	for rows.Next() {
		var j model.TranscodeJob
		if err := rows.Scan(&j.ID, &j.TrackID, &j.Status, &j.Progress, &j.Error, &j.Attempts, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *Repos) Audit(ctx context.Context, actor, action, reason string, meta any) {
	raw, _ := json.Marshal(meta)
	if raw == nil {
		raw = []byte("{}")
	}
	_, _ = r.Pool.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_id, action, reason, meta, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), actor, action, reason, raw, clock.Now())
}

func (r *Repos) ListAudit(ctx context.Context, limit int) ([]model.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.Pool.Query(ctx, `
		SELECT id, actor_id, action, reason, meta::text, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AuditLog{}
	for rows.Next() {
		var a model.AuditLog
		if err := rows.Scan(&a.ID, &a.ActorID, &a.Action, &a.Reason, &a.Meta, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repos) CreateUpload(ctx context.Context, u *model.UploadSession) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO upload_sessions (id, user_id, filename, sha256, size_bytes, chunk_size, received, tmp_key, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		u.ID, u.UserID, u.Filename, u.SHA256, u.SizeBytes, u.ChunkSize, u.Received, u.TmpKey, u.CreatedAt)
	return err
}

func (r *Repos) UploadByID(ctx context.Context, id string) (*model.UploadSession, error) {
	row := r.Pool.QueryRow(ctx, `
		SELECT id, user_id, filename, sha256, size_bytes, chunk_size, received, tmp_key, created_at
		FROM upload_sessions WHERE id=$1`, id)
	var u model.UploadSession
	if err := row.Scan(&u.ID, &u.UserID, &u.Filename, &u.SHA256, &u.SizeBytes, &u.ChunkSize, &u.Received, &u.TmpKey, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repos) MarkChunk(ctx context.Context, id string, idx int) error {
	_, err := r.Pool.Exec(ctx, `UPDATE upload_sessions SET received[$1]=true WHERE id=$2`, idx+1, id)
	return err
}

func (r *Repos) Stats(ctx context.Context) (map[string]any, error) {
	var users, tracks, orders int
	_ = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&users)
	_ = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&tracks)
	_ = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE status='paid'`).Scan(&orders)
	return map[string]any{"users": users, "tracks": tracks, "paidOrders": orders}, nil
}

func NewID() string { return uuid.NewString() }

func NewOrderNo() string {
	return fmt.Sprintf("GM%s", strings.ReplaceAll(uuid.NewString(), "-", "")[:16])
}
