package service

import "gomusical/internal/model"

// Matrix documents the three-tier access union. Used by tests and Decide().
type MatrixRow struct {
	Preview  bool
	FullPlay bool
	Download bool
}

func PermissionMatrix(tr *model.Track, isOwner, isFan, hasPaid bool) MatrixRow {
	row := MatrixRow{Preview: true}
	if isOwner {
		return MatrixRow{Preview: true, FullPlay: true, Download: true}
	}
	if hasPaid {
		row.FullPlay = true
		if tr.PaidDownload {
			row.Download = true
		}
	}
	if isFan {
		if tr.FanOnly || !tr.FanOnly {
			row.FullPlay = true
		}
		if tr.FanDownload {
			row.Download = true
		}
	}
	return row
}

func TrackReady(tr *model.Track) bool {
	return tr != nil && tr.TranscodeStatus == model.JobReady && tr.SegmentCount > 0
}
