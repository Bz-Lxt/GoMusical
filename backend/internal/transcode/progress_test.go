package transcode

import (
	"testing"
	"time"
)

func TestParseFFmpegTime(t *testing.T) {
	d, ok := ParseFFmpegTime("frame=1\ntime=00:00:12.50 bitrate=128")
	if !ok || d != 12500*time.Millisecond {
		t.Fatalf("%v %v", d, ok)
	}
}

func TestProgressPercent(t *testing.T) {
	if ProgressPercent(18*time.Second, 36*time.Second) != 50 {
		t.Fatal("50")
	}
	if ProgressPercent(40*time.Second, 36*time.Second) != 99 {
		t.Fatal("cap 99 until done")
	}
}

func TestClassify(t *testing.T) {
	if ClassifyFFmpegError("Invalid data found when processing") != "invalid_media" {
		t.Fatal("class")
	}
}
