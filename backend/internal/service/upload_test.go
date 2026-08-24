package service

import (
	"testing"

	"gomusical/internal/model"
)

func TestPlanInstant(t *testing.T) {
	sha := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	p, err := PlanUpload(1024, "a.wav", sha, &model.AssetBlob{SHA256: sha})
	if err != nil || !p.Instant {
		t.Fatalf("%v %+v", err, p)
	}
}

func TestPlanChunks(t *testing.T) {
	sha := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	p, err := PlanUpload(12*1024*1024, "song.flac", sha, nil)
	if err != nil || p.Chunks != 3 {
		t.Fatalf("%v %+v", err, p)
	}
}

func TestRejectExe(t *testing.T) {
	sha := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if _, err := PlanUpload(10, "x.exe", sha, nil); err == nil {
		t.Fatal("exe")
	}
}

func TestRecvComplete(t *testing.T) {
	if RecvComplete([]bool{true, false}) || !RecvComplete([]bool{true, true}) {
		t.Fatal("recv")
	}
	if len(MissingChunks([]bool{true, false, true})) != 1 {
		t.Fatal("missing")
	}
}

func TestDisplayTitle(t *testing.T) {
	if DisplayTitle("  河  ", "x.wav") != "河" && DisplayTitle("", "河对岸的灯.wav") != "河对岸的灯" {
		if DisplayTitle("", "河对岸的灯.wav") != "河对岸的灯" {
			t.Fatal(DisplayTitle("", "河对岸的灯.wav"))
		}
	}
}
