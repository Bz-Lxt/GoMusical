package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(addr, pass string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
}

func Ping(ctx context.Context, c *redis.Client) error {
	if c == nil {
		return fmt.Errorf("redis nil")
	}
	return c.Ping(ctx).Err()
}
