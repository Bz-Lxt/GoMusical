package hmacx

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func NormalizeFingerprint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "anon"
	}
	if len(raw) > 240 {
		raw = raw[:240]
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:8])
}

func FingerprintMatch(bound, got string) bool {
	if bound == "" || bound == "anon" {
		return true
	}
	return bound == NormalizeFingerprint(got) || bound == got
}
