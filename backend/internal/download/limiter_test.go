package download

import "testing"

func TestTokenBucketRefill(t *testing.T) {
	l := NewLimiter(nil, 2, 20, 8*1024*1024, 80*1024*1024)
	l.WaitBytes(1024)
	l.WaitUser("u1", 2048)
	if l.global.tokens < 0 {
		t.Fatal("tokens went negative")
	}
}

func TestCivilDayUsesBeijing(t *testing.T) {
	d := civilDay()
	if len(d) != 10 {
		t.Fatalf("day key %s", d)
	}
}
