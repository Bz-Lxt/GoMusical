package stream

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"gomusical/internal/hmacx"
)

// Revoker stores stream tokens that creators/admins have killed.
type Revoker struct {
	RDB *redis.Client
}

func (r *Revoker) Kill(ctx context.Context, token string, ttl time.Duration) error {
	if r == nil || r.RDB == nil {
		return nil
	}
	return r.RDB.Set(ctx, "stream:kill:"+token, "1", ttl).Err()
}

func (r *Revoker) Dead(ctx context.Context, token string) bool {
	if r == nil || r.RDB == nil {
		return false
	}
	n, err := r.RDB.Exists(ctx, "stream:kill:"+token).Result()
	return err == nil && n > 0
}

func (r *Revoker) Remember(ctx context.Context, token string, ss hmacx.StreamSession, ttl time.Duration) error {
	if r == nil || r.RDB == nil {
		return nil
	}
	b, err := json.Marshal(ss)
	if err != nil {
		return err
	}
	return r.RDB.Set(ctx, "stream:live:"+ss.TrackID+":"+ss.UserID, b, ttl).Err()
}
