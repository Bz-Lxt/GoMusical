package auth

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := HashPassword("Creator123!")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("Creator123!", h) {
		t.Fatal("should verify")
	}
	if VerifyPassword("wrong-pass", h) {
		t.Fatal("should reject")
	}
}

func TestShortPassword(t *testing.T) {
	if _, err := HashPassword("123"); err == nil {
		t.Fatal("short password must fail")
	}
}
