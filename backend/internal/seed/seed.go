package seed

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"gomusical/internal/auth"
	"gomusical/internal/clock"
	"gomusical/internal/logx"
	"gomusical/internal/model"
	"gomusical/internal/repo"
	"gomusical/internal/storage"
	"gomusical/internal/transcode"
)

func Run(ctx context.Context, repos *repo.Repos, store *storage.Local, eng *transcode.Engine, worker *transcode.Worker, ffmpeg string) error {
	if _, err := repos.UserByEmail(ctx, "creator@gomusical.local"); err == nil {
		logx.Info("seed skipped, users exist")
		return nil
	}
	mk := func(email, pass, name, role, bio string) *model.User {
		h, _ := auth.HashPassword(pass)
		u := &model.User{
			ID: uuid.NewString(), Email: email, PasswordHash: h,
			DisplayName: name, Role: role, Bio: bio, CreatedAt: clock.Now(),
		}
		_ = repos.CreateUser(ctx, u)
		return u
	}
	creator := mk("creator@gomusical.local", "Creator123!", "林晚舟", model.RoleCreator, "独立民谣 / 深夜播客。把没有厂牌的声音放上货架。")
	listener := mk("listener@gomusical.local", "Listener123!", "阿北", model.RoleListener, "随身带着耳机的听众。")
	_ = mk("admin@gomusical.local", "Admin123!", "货架管理员", model.RoleAdmin, "")

	album := &model.Album{
		ID: uuid.NewString(), CreatorID: creator.ID, Title: "午夜练习曲",
		Description: "未发行的卧室录音。", CreatedAt: clock.Now(),
	}
	_ = repos.CreateAlbum(ctx, album)

	src := filepath.Join(store.Root, "tmp", "seed36.wav")
	_ = os.MkdirAll(filepath.Dir(src), 0o755)
	if err := exec.CommandContext(ctx, ffmpeg, "-y", "-f", "lavfi",
		"-i", "sine=frequency=440:sample_rate=44100:duration=36",
		"-c:a", "pcm_s16le", src).Run(); err != nil {
		logx.Error("seed ffmpeg wav", "err", err)
		return err
	}
	sha, size, err := storage.HashFile(src)
	if err != nil {
		return err
	}
	key := storage.BlobKey(sha, "wav")
	dst, _ := store.Resolve(key)
	_ = os.MkdirAll(filepath.Dir(dst), 0o755)
	_ = os.Rename(src, dst)
	_ = repos.UpsertBlob(ctx, &model.AssetBlob{SHA256: sha, StorageKey: key, SizeBytes: size, MIME: "audio/wav", CreatedAt: clock.Now()})

	tr := &model.Track{
		ID: uuid.NewString(), CreatorID: creator.ID, AlbumID: &album.ID,
		Title: "河对岸的灯", DisplayFilename: "河对岸的灯.wav", DurationMS: 36000,
		Format: "wav", ContentSHA256: sha, StorageKey: key, SizeBytes: size,
		PreviewSeconds: 30, PaidDownload: true, PaidPriceCents: 900,
		FanOnly: false, FanDownload: false, TranscodeStatus: model.JobPending,
		CreatedAt: clock.Now(), UpdatedAt: clock.Now(),
	}
	if err := repos.CreateTrack(ctx, tr); err != nil {
		return err
	}
	if worker != nil {
		_, _ = worker.Enqueue(ctx, tr.ID)
	}

	c := &model.Comment{
		ID: uuid.NewString(), TrackID: tr.ID, UserID: listener.ID,
		TimestampMS: 12000, Body: "12 秒那个泛音像有人推开窗。", CreatedAt: clock.Now(),
	}
	_ = repos.CreateComment(ctx, c)
	logx.Info("seed ready", "track", tr.ID, "creator", creator.Email)
	return nil
}
