package hmacx

import "testing"

func TestFingerprintStable(t *testing.T) {
	a := NormalizeFingerprint("Mozilla/5.0")
	b := NormalizeFingerprint("Mozilla/5.0")
	if a != b || a == "anon" {
		t.Fatalf("%s %s", a, b)
	}
	if !FingerprintMatch(a, "Mozilla/5.0") {
		t.Fatal("match")
	}
	if FingerprintMatch(a, "Other") {
		t.Fatal("mismatch")
	}
}
