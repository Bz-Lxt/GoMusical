package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

const (
	ChunkBytes   = 5 * 1024 * 1024
	MaxUpload    = 500 * 1024 * 1024
	AllowedExt   = ".flac,.wav,.mp3,.m4a,.aac"
)

type UploadPlan struct {
	Chunks    int
	ChunkSize int
	Instant   bool
}

func PlanUpload(size int64, filename, sha string, existing *model.AssetBlob) (UploadPlan, error) {
	if size <= 0 || size > MaxUpload {
		return UploadPlan{}, httpx.New(400, "bad_request", "文件大小超出 1B–500MB")
	}
	if !httpx.SHA256Hex(sha) {
		return UploadPlan{}, httpx.New(400, "bad_request", "sha256 非法")
	}
	if !httpx.Filename(filename) || !AllowedAudio(filename) {
		return UploadPlan{}, httpx.New(400, "bad_request", "仅接受 FLAC/WAV/MP3/M4A/AAC")
	}
	if existing != nil && existing.SHA256 == sha {
		return UploadPlan{Instant: true}, nil
	}
	n := int((size + ChunkBytes - 1) / ChunkBytes)
	if n < 1 {
		n = 1
	}
	return UploadPlan{Chunks: n, ChunkSize: ChunkBytes}, nil
}

func AllowedAudio(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, a := range strings.Split(AllowedExt, ",") {
		if ext == a {
			return true
		}
	}
	return false
}

func RecvComplete(flags []bool) bool {
	if len(flags) == 0 {
		return false
	}
	for _, f := range flags {
		if !f {
			return false
		}
	}
	return true
}

func MissingChunks(flags []bool) []int {
	out := []int{}
	for i, f := range flags {
		if !f {
			out = append(out, i)
		}
	}
	return out
}

func DisplayTitle(title, filename string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func FormatFromName(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	switch ext {
	case "flac", "wav", "mp3", "aac", "m4a":
		return ext
	default:
		return "bin"
	}
}

func AssertHash(got, want string, gotSize, wantSize int64) error {
	if !strings.EqualFold(got, want) || gotSize != wantSize {
		return httpx.New(422, "hash_mismatch", fmt.Sprintf("申报 %d/%s 实际 %d/%s", wantSize, want[:8], gotSize, got[:8]))
	}
	return nil
}
