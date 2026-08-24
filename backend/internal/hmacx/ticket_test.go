package hmacx

import (
	"strings"
	"testing"
	"time"
)

func TestSignVerifyTicket(t *testing.T) {
	secret := []byte("unit-test-secret-32bytes-minimum")
	tk := NewTicket("tr1", "u1", "g1", "lossless", 5*time.Minute, 3)
	s, err := SignTicket(tk, secret)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(s, ".") {
		t.Fatalf("ticket must not contain dot: %s", s)
	}
	got, err := VerifyTicket(s, secret)
	if err != nil {
		t.Fatal(err)
	}
	if got.TrackID != "tr1" || got.Nonce != tk.Nonce {
		t.Fatalf("mismatch %+v", got)
	}
}

func TestTamperTicket(t *testing.T) {
	secret := []byte("unit-test-secret-32bytes-minimum")
	tk := NewTicket("tr1", "u1", "g1", "lossless", time.Minute, 3)
	s, _ := SignTicket(tk, secret)
	bad := s[:len(s)-2] + "aa"
	if _, err := VerifyTicket(bad, secret); err == nil {
		t.Fatal("expected tamper error")
	}
}

func TestExpiredTicket(t *testing.T) {
	secret := []byte("unit-test-secret-32bytes-minimum")
	tk := NewTicket("tr1", "u1", "g1", "lossless", -time.Second, 3)
	s, _ := SignTicket(tk, secret)
	if _, err := VerifyTicket(s, secret); err == nil {
		t.Fatal("expected expired")
	}
}

func TestStreamSign(t *testing.T) {
	secret := []byte("unit-test-secret-32bytes-minimum")
	ss := NewStream("u1", "tr1", "PREVIEW", "fp", 30000, time.Minute)
	tok, err := SignStream(ss, secret)
	if err != nil {
		t.Fatal(err)
	}
	got, err := VerifyStream(tok, secret)
	if err != nil {
		t.Fatal(err)
	}
	if got.UntilMS != 30000 {
		t.Fatalf("until %d", got.UntilMS)
	}
}

func TestMaxSegmentIndex(t *testing.T) {
	if MaxSegmentIndex(30000, 6000) != 4 {
		t.Fatalf("want 4 got %d", MaxSegmentIndex(30000, 6000))
	}
	if MaxSegmentIndex(36000, 6000) != 5 {
		t.Fatalf("want 5 got %d", MaxSegmentIndex(36000, 6000))
	}
}
