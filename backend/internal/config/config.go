package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env            string
	HTTPAddr       string
	PublicOrigin   string
	AllowedOrigins []string
	RefererAllow   []string

	DatabaseURL string
	RedisAddr   string
	RedisPass   string

	StorageRoot string
	FFmpegBin   string
	FFprobeBin  string

	HMACSecret     []byte
	SessionSecret  []byte
	SessionTTL     time.Duration
	StreamTTL      time.Duration
	TicketTTL      time.Duration
	TicketTTLMax   time.Duration
	TicketMaxUses  int
	DownloadConc   int
	DailyDownloads int
	UserBPS        int64
	GlobalBPS      int64

	PaymentMode     string
	PaymentRealKey  string
	MockPayBehavior string

	SeedEnabled bool
	LogLevel    string
}

func Load() Config {
	secret := getenv("HMAC_SECRET", "gomusical-dev-hmac-secret-change-me-32b")
	sess := getenv("SESSION_SECRET", "gomusical-dev-session-secret-change-me")
	origins := splitCSV(getenv("CORS_ORIGINS", "http://localhost:29471"))
	cfg := Config{
		Env:             getenv("APP_ENV", "development"),
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		PublicOrigin:    getenv("PUBLIC_ORIGIN", "http://localhost:29471"),
		AllowedOrigins:  origins,
		RefererAllow:    splitCSV(getenv("REFERER_ALLOW", "http://localhost:29471")),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://gomusical:gomusical@localhost:5432/gomusical?sslmode=disable"),
		RedisAddr:       getenv("REDIS_ADDR", "localhost:6379"),
		RedisPass:       getenv("REDIS_PASSWORD", ""),
		StorageRoot:     getenv("STORAGE_ROOT", "/data/storage"),
		FFmpegBin:       getenv("FFMPEG_BIN", "ffmpeg"),
		FFprobeBin:      getenv("FFPROBE_BIN", "ffprobe"),
		HMACSecret:      []byte(secret),
		SessionSecret:   []byte(sess),
		SessionTTL:      durationHours("SESSION_TTL_HOURS", 72),
		StreamTTL:       durationSeconds("STREAM_TTL_SEC", 600),
		TicketTTL:       durationSeconds("TICKET_TTL_SEC", 300),
		TicketTTLMax:    durationSeconds("TICKET_TTL_MAX_SEC", 900),
		TicketMaxUses:   intEnv("TICKET_MAX_USES", 3),
		DownloadConc:    intEnv("DOWNLOAD_CONCURRENCY", 2),
		DailyDownloads:  intEnv("DAILY_DOWNLOAD_LIMIT", 20),
		UserBPS:         int64Env("USER_DOWNLOAD_BPS", 8*1024*1024),
		GlobalBPS:       int64Env("GLOBAL_DOWNLOAD_BPS", 80*1024*1024),
		PaymentMode:     strings.ToLower(getenv("PAYMENT_MODE", "mock")),
		PaymentRealKey:  getenv("PAYMENT_REAL_KEY", ""),
		MockPayBehavior: getenv("MOCK_PAY_BEHAVIOR", "success"),
		SeedEnabled:     getenv("SEED_ENABLED", "true") == "true",
		LogLevel:        getenv("LOG_LEVEL", "info"),
	}
	if cfg.PaymentMode == "real" && strings.TrimSpace(cfg.PaymentRealKey) == "" {
		cfg.PaymentMode = "mock"
	}
	return cfg
}

func (c Config) Validate() error {
	if len(c.HMACSecret) < 16 {
		return fmt.Errorf("HMAC_SECRET too short")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL required")
	}
	return nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func intEnv(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func int64Env(k string, def int64) int64 {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return def
}

func durationSeconds(k string, def int) time.Duration {
	return time.Duration(intEnv(k, def)) * time.Second
}

func durationHours(k string, def int) time.Duration {
	return time.Duration(intEnv(k, def)) * time.Hour
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
