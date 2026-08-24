package download

import "testing"

func TestParseSingleRange(t *testing.T) {
	start, end, ok := ParseSingleRange("bytes=0-1023", 5000)
	if !ok || start != 0 || end != 1023 {
		t.Fatalf("%d %d %v", start, end, ok)
	}
	start, end, ok = ParseSingleRange("bytes=1000-", 5000)
	if !ok || start != 1000 || end != 4999 {
		t.Fatalf("open end %d %d", start, end)
	}
	start, end, ok = ParseSingleRange("bytes=-100", 5000)
	if !ok || start != 4900 || end != 4999 {
		t.Fatalf("suffix %d %d", start, end)
	}
	if _, _, ok := ParseSingleRange("bytes=0-10,20-30", 5000); ok {
		t.Fatal("multi range should be rejected for resume accounting")
	}
}

func TestBytesOfRange(t *testing.T) {
	if n := BytesOfRange("bytes=0-1023", 5000); n != 1024 {
		t.Fatalf("got %d", n)
	}
}
