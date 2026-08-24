package service

import (
	"context"

	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
	"gomusical/internal/repo"
	"gomusical/internal/stream"
)

type Access struct {
	Repos *repo.Repos
}

type Decision struct {
	Tier          string
	UntilMS       int
	CanDownload   bool
	GrantID       string
	PreviewOnly   bool
	IsFan         bool
	Paid          bool
	MaxSegIndex   int
}

func (a *Access) Decide(ctx context.Context, userID string, tr *model.Track) Decision {
	d := Decision{
		Tier:        model.TierPreview,
		UntilMS:     tr.PreviewSeconds * 1000,
		PreviewOnly: true,
	}
	if d.UntilMS <= 0 {
		d.UntilMS = 30000
	}
	if userID != "" && userID == tr.CreatorID {
		d.Tier = model.TierPaid
		d.UntilMS = tr.DurationMS
		d.CanDownload = true
		d.PreviewOnly = false
		d.Paid = true
		d.MaxSegIndex = tr.SegmentCount - 1
		return d
	}
	if userID != "" {
		if sub, err := a.Repos.ActiveSub(ctx, userID, tr.CreatorID); err == nil && sub != nil {
			d.IsFan = true
			if tr.FanOnly || true {
				d.Tier = model.TierFan
				d.UntilMS = tr.DurationMS
				d.PreviewOnly = false
			}
			if tr.FanDownload {
				d.CanDownload = true
			}
		}
		grants, _ := a.Repos.GrantsFor(ctx, userID, tr.ID)
		now := clock.Now()
		for _, g := range grants {
			if g.ExpiresAt != nil && !g.ExpiresAt.After(now) {
				continue
			}
			if g.Kind == model.TierPaid || g.Kind == "FAN_DOWNLOAD" {
				d.Paid = true
				d.CanDownload = true
				d.GrantID = g.ID
				d.Tier = model.TierPaid
				d.UntilMS = tr.DurationMS
				d.PreviewOnly = false
			}
		}
	}
	if tr.DurationMS > 0 && d.UntilMS > tr.DurationMS {
		d.UntilMS = tr.DurationMS
	}
	d.MaxSegIndex = stream.SegmentDurationMS
	// compute max segment from until
	if d.UntilMS <= 0 {
		d.MaxSegIndex = -1
	} else {
		d.MaxSegIndex = (d.UntilMS + stream.SegmentDurationMS - 1) / stream.SegmentDurationMS - 1
	}
	if tr.SegmentCount > 0 && d.MaxSegIndex >= tr.SegmentCount {
		d.MaxSegIndex = tr.SegmentCount - 1
	}
	return d
}

func (a *Access) AssertCommentWindow(ctx context.Context, userID string, tr *model.Track, ts int) error {
	if ts < 0 {
		return httpx.ErrCommentWindow
	}
	d := a.Decide(ctx, userID, tr)
	if ts > d.UntilMS {
		return httpx.ErrCommentWindow
	}
	return nil
}
