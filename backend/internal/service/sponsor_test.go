package service

import (
	"testing"

	"gomusical/internal/hmacx"
	"time"
)

func TestTicketRoundTripForSponsor(t *testing.T) {
	secret := []byte("unit-test-secret-32bytes-minimum")
	tk := hmacx.NewTicket("t", "u", "g", "lossless", time.Minute, 3)
	s, err := hmacx.SignTicket(tk, secret)
	if err != nil {
		t.Fatal(err)
	}
	got, err := hmacx.VerifyTicket(s, secret)
	if err != nil || got.GrantID != "g" {
		t.Fatalf("%v %+v", err, got)
	}
}
