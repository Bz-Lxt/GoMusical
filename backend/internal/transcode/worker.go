package transcode

import (
	"context"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gomusical/internal/clock"
	"gomusical/internal/logx"
	"gomusical/internal/model"
	"gomusical/internal/repo"
	"gomusical/internal/storage"
)

type Worker struct {
	Eng   *Engine
	Repos *repo.Repos
	Store *storage.Local
	stop  chan struct{}
}

func (w *Worker) Start(ctx context.Context) {
	w.stop = make(chan struct{})
	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stop:
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *Worker) Stop() {
	if w.stop != nil {
		close(w.stop)
	}
}

func (w *Worker) Enqueue(ctx context.Context, trackID string) (*model.TranscodeJob, error) {
	j := &model.TranscodeJob{
		ID: uuid.NewString(), TrackID: trackID, Status: model.JobPending,
		CreatedAt: clock.Now(), UpdatedAt: clock.Now(),
	}
	if err := w.Repos.CreateJob(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}

func (w *Worker) tick(ctx context.Context) {
	jobs, err := w.Repos.PendingJobs(ctx)
	if err != nil {
		logx.Error("list pending jobs", "err", err)
		return
	}
	for _, j := range jobs {
		w.runOne(ctx, j)
	}
}

func (w *Worker) runOne(ctx context.Context, j model.TranscodeJob) {
	j.Status = model.JobRunning
	j.Attempts++
	j.UpdatedAt = clock.Now()
	_ = w.Repos.UpdateJob(ctx, &j)

	tr, err := w.Repos.TrackByID(ctx, j.TrackID)
	if err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}
	src, err := w.Store.Resolve(tr.StorageKey)
	if err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}

	// Content-hash cache: if another track already produced HLS for same sha, copy keys.
	if tr.ContentSHA256 != "" {
		siblings, _ := w.Repos.ListTracks(ctx, "", "")
		for _, s := range siblings {
			if s.ID != tr.ID && s.ContentSHA256 == tr.ContentSHA256 && s.TranscodeStatus == model.JobReady && s.HLSDir != "" {
				tr.HLSDir = s.HLSDir
				tr.PeaksKey = s.PeaksKey
				tr.SegmentCount = s.SegmentCount
				tr.DurationMS = s.DurationMS
				tr.TranscodeStatus = model.JobReady
				tr.UpdatedAt = clock.Now()
				_ = w.Repos.UpdateTrack(ctx, tr)
				j.Status = model.JobReady
				j.Progress = 100
				j.UpdatedAt = clock.Now()
				_ = w.Repos.UpdateJob(ctx, &j)
				logx.Info("transcode cache hit", "sha", tr.ContentSHA256, "track", tr.ID)
				return
			}
		}
	}

	probe, err := w.Eng.Probe(ctx, src)
	if err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}
	tr.DurationMS = probe.DurationMS
	if tr.Format == "" {
		tr.Format = probe.Format
	}
	j.Progress = 20
	_ = w.Repos.UpdateJob(ctx, &j)

	hlsKey := storage.HLSDir(tr.ID)
	hlsAbs, _ := w.Store.Resolve(hlsKey)
	n, err := w.Eng.HLS(ctx, src, filepath.Join(hlsAbs, "256k"), "256k")
	if err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}
	if _, err := w.Eng.HLS(ctx, src, filepath.Join(hlsAbs, "128k"), "128k"); err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}
	j.Progress = 70
	_ = w.Repos.UpdateJob(ctx, &j)

	peaksKey := storage.PeaksKey(tr.ID)
	peaksAbs, _ := w.Store.Resolve(peaksKey)
	if err := w.Eng.Peaks(ctx, src, peaksAbs, 8000); err != nil {
		w.fail(ctx, &j, err.Error())
		return
	}
	coverKey := storage.CoverKey(tr.ID)
	coverAbs, _ := w.Store.Resolve(coverKey)
	_ = w.Eng.Cover(ctx, src, coverAbs)

	tr.HLSDir = hlsKey
	tr.PeaksKey = peaksKey
	tr.CoverKey = coverKey
	tr.SegmentCount = n
	tr.TranscodeStatus = model.JobReady
	tr.TranscodeError = ""
	tr.UpdatedAt = clock.Now()
	_ = w.Repos.UpdateTrack(ctx, tr)

	j.Status = model.JobReady
	j.Progress = 100
	j.Error = ""
	j.UpdatedAt = clock.Now()
	_ = w.Repos.UpdateJob(ctx, &j)
	logx.Info("transcode ready", "track", tr.ID, "segments", n, "durationMs", tr.DurationMS)
}

func (w *Worker) fail(ctx context.Context, j *model.TranscodeJob, msg string) {
	j.Status = model.JobFailed
	j.Error = msg
	j.UpdatedAt = clock.Now()
	_ = w.Repos.UpdateJob(ctx, j)
	if tr, err := w.Repos.TrackByID(ctx, j.TrackID); err == nil {
		tr.TranscodeStatus = model.JobFailed
		tr.TranscodeError = msg
		tr.UpdatedAt = clock.Now()
		_ = w.Repos.UpdateTrack(ctx, tr)
	}
	logx.Error("transcode failed", "job", j.ID, "err", msg)
}
