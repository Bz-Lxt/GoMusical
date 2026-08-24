package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
	"gomusical/internal/model"
	"gomusical/internal/repo"
)

type Grants struct {
	Repos *repo.Repos
}

func (g *Grants) EnsurePaid(ctx context.Context, userID, trackID string) (*model.Grant, error) {
	existing, err := g.Repos.GrantsFor(ctx, userID, trackID)
	if err != nil {
		return nil, err
	}
	now := clock.Now()
	for i := range existing {
		if existing[i].Kind == model.TierPaid && (existing[i].ExpiresAt == nil || existing[i].ExpiresAt.After(now)) {
			return &existing[i], nil
		}
	}
	grant := &model.Grant{
		ID: uuid.NewString(), UserID: userID, TrackID: &trackID,
		Kind: model.TierPaid, CreatedAt: now,
	}
	if err := g.Repos.CreateGrant(ctx, grant); err != nil {
		return nil, err
	}
	return grant, nil
}

func (g *Grants) ActiveKinds(ctx context.Context, userID, trackID string) ([]string, error) {
	list, err := g.Repos.GrantsFor(ctx, userID, trackID)
	if err != nil {
		return nil, err
	}
	now := clock.Now()
	out := []string{}
	for _, x := range list {
		if x.ExpiresAt != nil && !x.ExpiresAt.After(now) {
			continue
		}
		out = append(out, x.Kind)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (g *Grants) RevokeKind(ctx context.Context, userID, trackID, kind string) error {
	// Expiry-based revoke: write a grant that already expired so Decide() ignores it.
	// Historical rows stay for audit; new Decide() reads unexpired only.
	past := clock.Now().Add(-time.Second)
	dummy := &model.Grant{
		ID: uuid.NewString(), UserID: userID, TrackID: &trackID,
		Kind: kind, ExpiresAt: &past, CreatedAt: clock.Now(),
	}
	if err := g.Repos.CreateGrant(ctx, dummy); err != nil {
		return httpx.Wrap(500, "internal", "无法写入吊销记录", err)
	}
	g.Repos.Audit(ctx, userID, "grant.revoke", kind, map[string]any{"track": trackID})
	return nil
}
