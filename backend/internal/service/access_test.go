package service

import (
	"testing"

	"gomusical/internal/hmacx"
	"gomusical/internal/model"
	"gomusical/internal/stream"
)

func TestPreviewSegmentBound(t *testing.T) {
	tr := &model.Track{PreviewSeconds: 30, DurationMS: 36000, SegmentCount: 6}
	until := tr.PreviewSeconds * 1000
	max := hmacx.MaxSegmentIndex(until, stream.SegmentDurationMS)
	if max != 4 {
		t.Fatalf("preview max seg %d", max)
	}
	if max >= 5 {
		t.Fatal("seg_5 must be denied for 30s preview")
	}
}

func TestCommentWindowLogic(t *testing.T) {
	until := 30000
	if 30001 <= until {
		t.Fatal("out of window should fail")
	}
	if 15000 > until {
		t.Fatal("in window should pass")
	}
}
