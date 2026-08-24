package service

import (
	"testing"

	"gomusical/internal/model"
)

func TestPermissionOwner(t *testing.T) {
	tr := &model.Track{PaidDownload: true}
	m := PermissionMatrix(tr, true, false, false)
	if !m.Download || !m.FullPlay {
		t.Fatal("owner must have full rights")
	}
}

func TestPermissionPreviewOnly(t *testing.T) {
	tr := &model.Track{PaidDownload: true, FanOnly: false}
	m := PermissionMatrix(tr, false, false, false)
	if m.FullPlay || m.Download {
		t.Fatal("anonymous should be preview only")
	}
}

func TestPermissionPaid(t *testing.T) {
	tr := &model.Track{PaidDownload: true}
	m := PermissionMatrix(tr, false, false, true)
	if !m.Download || !m.FullPlay {
		t.Fatal("paid grant should unlock download")
	}
}

func TestPermissionFanDownloadOff(t *testing.T) {
	tr := &model.Track{FanDownload: false, PaidDownload: true}
	m := PermissionMatrix(tr, false, true, false)
	if !m.FullPlay {
		t.Fatal("fan should hear full track")
	}
	if m.Download {
		t.Fatal("fan download switch off")
	}
}
