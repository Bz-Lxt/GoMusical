package hmacx

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gomusical/internal/clock"
	"gomusical/internal/httpx"
)

// Ticket is a handwritten HMAC credential. URL-safe encoding uses '~' not '.'
// because chi /:ticket would treat '.' as a file extension (see knowledge-base).
type Ticket struct {
	TrackID string `json:"trackId"`
	UserID  string `json:"userId"`
	GrantID string `json:"grantId"`
	Exp     int64  `json:"exp"`
	MaxUses int    `json:"maxUses"`
	Nonce   string `json:"nonce"`
	Scope   string `json:"scope"`
}

func NewTicket(trackID, userID, grantID, scope string, ttl time.Duration, maxUses int) Ticket {
	return Ticket{
		TrackID: trackID,
		UserID:  userID,
		GrantID: grantID,
		Exp:     clock.Now().Add(ttl).Unix(),
		MaxUses: maxUses,
		Nonce:   randomNonce(16),
		Scope:   scope,
	}
}

func SignTicket(t Ticket, secret []byte) (string, error) {
	if t.TrackID == "" || t.UserID == "" || t.GrantID == "" || t.Nonce == "" || t.Scope == "" {
		return "", httpx.ErrBadRequest
	}
	if t.MaxUses <= 0 {
		return "", httpx.Wrap(400, "bad_request", "maxUses 必须为正整数", nil)
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "~" + sig, nil
}

func VerifyTicket(token string, secret []byte) (Ticket, error) {
	var zero Ticket
	parts := strings.Split(token, "~")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return zero, httpx.ErrTicketTampered
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0]))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return zero, httpx.ErrTicketTampered
	}
	if !hmac.Equal(want, got) {
		return zero, httpx.ErrTicketTampered
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return zero, httpx.ErrTicketTampered
	}
	var t Ticket
	if err := json.Unmarshal(raw, &t); err != nil {
		return zero, httpx.ErrTicketTampered
	}
	if t.TrackID == "" || t.UserID == "" || t.GrantID == "" || t.Nonce == "" {
		return zero, httpx.ErrTicketTampered
	}
	if t.Exp <= clock.Now().Unix() {
		return zero, httpx.ErrTicketExpired
	}
	return t, nil
}

func randomNonce(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", clock.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
