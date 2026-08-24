package download

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gomusical/internal/clock"
	"gomusical/internal/httpx"
)

// Limiter enforces concurrency, daily count, and token-bucket throughput.
// Redis is the source of truth for live counters; Postgres holds audit only.
type Limiter struct {
	RDB        *redis.Client
	MaxConc    int
	DailyLimit int
	UserBPS    int64
	GlobalBPS  int64

	mu     sync.Mutex
	users  map[string]*bucket
	global *bucket
}

type bucket struct {
	rate  float64
	burst float64
	tokens float64
	last  time.Time
}

func NewLimiter(rdb *redis.Client, conc, daily int, userBPS, globalBPS int64) *Limiter {
	return &Limiter{
		RDB:        rdb,
		MaxConc:    conc,
		DailyLimit: daily,
		UserBPS:    userBPS,
		GlobalBPS:  globalBPS,
		users:      map[string]*bucket{},
		global:     &bucket{rate: float64(globalBPS), burst: float64(globalBPS), tokens: float64(globalBPS), last: clock.Now()},
	}
}

func (l *Limiter) Acquire(ctx context.Context, userID string) (func(), error) {
	if l.RDB == nil {
		return func() {}, nil
	}
	key := "dl:conc:" + userID
	n, err := l.RDB.Incr(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	_ = l.RDB.Expire(ctx, key, 30*time.Minute)
	if int(n) > l.MaxConc {
		_, _ = l.RDB.Decr(ctx, key).Result()
		return nil, httpx.ErrTooMany
	}
	dayKey := "dl:day:" + userID + ":" + civilDay()
	used, err := l.RDB.Get(ctx, dayKey).Int()
	if err != nil && err != redis.Nil {
		_, _ = l.RDB.Decr(ctx, key).Result()
		return nil, err
	}
	if used >= l.DailyLimit {
		_, _ = l.RDB.Decr(ctx, key).Result()
		return nil, httpx.New(429, "daily_limit", "今日下载次数已达上限")
	}
	release := func() {
		_, _ = l.RDB.Decr(context.Background(), key).Result()
	}
	return release, nil
}

func (l *Limiter) MarkComplete(ctx context.Context, userID string) {
	if l.RDB == nil {
		return
	}
	dayKey := "dl:day:" + userID + ":" + civilDay()
	pipe := l.RDB.TxPipeline()
	pipe.Incr(ctx, dayKey)
	pipe.Expire(ctx, dayKey, 36*time.Hour)
	_, _ = pipe.Exec(ctx)
}

func (l *Limiter) WaitBytes(n int) {
	if n <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := clock.Now()
	l.refill(l.global, now)
	need := float64(n)
	if l.global.tokens < need {
		deficit := need - l.global.tokens
		sleep := time.Duration(deficit/l.global.rate*float64(time.Second)) + time.Millisecond
		l.mu.Unlock()
		time.Sleep(sleep)
		l.mu.Lock()
		now = clock.Now()
		l.refill(l.global, now)
	}
	l.global.tokens -= need
	if l.global.tokens < 0 {
		l.global.tokens = 0
	}
}

func (l *Limiter) WaitUser(userID string, n int) {
	if n <= 0 {
		return
	}
	l.mu.Lock()
	b := l.users[userID]
	if b == nil {
		b = &bucket{rate: float64(l.UserBPS), burst: float64(l.UserBPS), tokens: float64(l.UserBPS), last: clock.Now()}
		l.users[userID] = b
	}
	now := clock.Now()
	l.refill(b, now)
	need := float64(n)
	if b.tokens < need && b.rate > 0 {
		deficit := need - b.tokens
		sleep := time.Duration(deficit/b.rate*float64(time.Second)) + time.Millisecond
		l.mu.Unlock()
		time.Sleep(sleep)
		l.mu.Lock()
		l.refill(b, clock.Now())
	}
	b.tokens -= need
	if b.tokens < 0 {
		b.tokens = 0
	}
	l.mu.Unlock()
}

func (l *Limiter) refill(b *bucket, now time.Time) {
	if b.rate <= 0 {
		return
	}
	elapsed := now.Sub(b.last).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.last = now
}

func (l *Limiter) Concurrent(ctx context.Context, userID string) int {
	if l.RDB == nil {
		return 0
	}
	n, _ := l.RDB.Get(ctx, "dl:conc:"+userID).Int()
	return n
}

func (l *Limiter) DailyUsed(ctx context.Context, userID string) int {
	if l.RDB == nil {
		return 0
	}
	n, _ := l.RDB.Get(ctx, "dl:day:"+userID+":"+civilDay()).Int()
	return n
}

func (l *Limiter) FlagAbuse(ctx context.Context, ip, reason string) {
	if l.RDB == nil {
		return
	}
	_ = l.RDB.Incr(ctx, "dl:abuse:"+ip+":"+reason).Err()
	_ = l.RDB.Expire(ctx, "dl:abuse:"+ip+":"+reason, 24*time.Hour).Err()
}

func civilDay() string {
	y, m, d := clock.CivilDate(clock.Now())
	return fmt.Sprintf("%04d-%02d-%02d", y, int(m), d)
}
