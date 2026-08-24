package hmacx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"gomusical/internal/clock"
	"gomusical/internal/httpx"
)

// StreamSession binds a short-lived HLS playlist to a user + track + tier.
// Token alphabet is [A-Za-z0-9_-] only — no '.' — so chi path params stay intact.
type StreamSession struct {
	UserID    string `json:"uid"`
	TrackID   string `json:"tid"`
	Tier      string `json:"tier"`
	UntilMS   int    `json:"until"`
	Exp       int64  `json:"exp"`
	Fingerprint string `json:"fp"`
	Nonce     string `json:"n"`
}

func NewStream(userID, trackID, tier, fp string, untilMS int, ttl time.Duration) StreamSession {
	return StreamSession{
		UserID:      userID,
		TrackID:     trackID,
		Tier:        tier,
		UntilMS:     untilMS,
		Exp:         clock.Now().Add(ttl).Unix(),
		Fingerprint: fp,
		Nonce:       randomNonce(8),
	}
}

func SignStream(s StreamSession, secret []byte) (string, error) {
	if s.TrackID == "" || s.Tier == "" {
		return "", httpx.ErrBadRequest
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "_" + sig, nil
}

func VerifyStream(token string, secret []byte) (StreamSession, error) {
	var zero StreamSession
	i := strings.LastIndex(token, "_")
	if i <= 0 || i == len(token)-1 {
		return zero, httpx.ErrUnauthorized
	}
	payload, sigHex := token[:i], token[i+1:]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	want := mac.Sum(nil)
	got, err := hex.DecodeString(sigHex)
	if err != nil || !hmac.Equal(want, got) {
		return zero, httpx.ErrUnauthorized
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return zero, httpx.ErrUnauthorized
	}
	var s StreamSession
	if err := json.Unmarshal(raw, &s); err != nil {
		return zero, httpx.ErrUnauthorized
	}
	if s.TrackID == "" || s.Exp <= clock.Now().Unix() {
		return zero, httpx.ErrGone
	}
	return s, nil
}

func MaxSegmentIndex(untilMS, segmentMS int) int {
	if segmentMS <= 0 {
		return 0
	}
	if untilMS <= 0 {
		return -1
	}
	// last included segment is the one that starts before untilMS
	n := (untilMS + segmentMS - 1) / segmentMS
	if n < 1 {
		return 0
	}
	return n - 1
}
