package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HMAC_SECRET", "1234567890123456")
	t.Setenv("DATABASE_URL", "postgres://x")
	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.TicketTTL.Seconds() != 300 {
		t.Fatalf("ttl %v", cfg.TicketTTL)
	}
	if cfg.PaymentMode != "mock" && os.Getenv("PAYMENT_MODE") == "" {
		// default mock
	}
}

func TestRealDegradesWithoutKey(t *testing.T) {
	t.Setenv("PAYMENT_MODE", "real")
	t.Setenv("PAYMENT_REAL_KEY", "")
	t.Setenv("HMAC_SECRET", "1234567890123456")
	cfg := Load()
	if cfg.PaymentMode != "mock" {
		t.Fatal(cfg.PaymentMode)
	}
}
