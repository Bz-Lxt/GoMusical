package stream

import "testing"

func TestInWindow(t *testing.T) {
	if !InWindow(4, 30000) {
		t.Fatal("seg 4 starts at 24s, should be in 30s window")
	}
	if InWindow(5, 30000) {
		t.Fatal("seg 5 starts at 30s, must be out")
	}
}

func TestAssertSafeToken(t *testing.T) {
	if err := AssertSafeToken("abc_def-123"); err != nil {
		t.Fatal(err)
	}
	if err := AssertSafeToken("../x"); err == nil {
		t.Fatal("path token")
	}
}
