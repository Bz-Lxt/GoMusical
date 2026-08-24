package stream

import (
	"strings"
	"testing"
)

func TestPlaylistPreviewTruncation(t *testing.T) {
	body := Playlist("/api/stream/tok", 6, 4, 6)
	if !strings.Contains(body, "seg_4.ts") {
		t.Fatal("missing last preview segment")
	}
	if strings.Contains(body, "seg_5.ts") {
		t.Fatal("preview playlist leaked full-track segment")
	}
	if !strings.Contains(body, "#EXT-X-ENDLIST") {
		t.Fatal("missing endlist")
	}
}

func TestParseSegName(t *testing.T) {
	n, ok := ParseSegName("seg_10.ts")
	if !ok || n != 10 {
		t.Fatalf("got %d %v", n, ok)
	}
	if _, ok := ParseSegName("../seg_1.ts"); ok {
		t.Fatal("path escape accepted")
	}
	if _, ok := ParseSegName("seg_x.ts"); ok {
		t.Fatal("non numeric accepted")
	}
}
