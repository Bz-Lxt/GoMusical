package transcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gomusical/internal/logx"
	"gomusical/internal/storage"
	"gomusical/internal/stream"
)

type Engine struct {
	FFmpeg  string
	FFprobe string
	Store   *storage.Local
}

type Probe struct {
	DurationMS int
	Format     string
	Codec      string
}

func (e *Engine) Probe(ctx context.Context, path string) (Probe, error) {
	cmd := exec.CommandContext(ctx, e.FFprobe, "-v", "error", "-show_entries", "format=duration,format_name:stream=codec_name", "-of", "json", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return Probe{}, fmt.Errorf("ffprobe: %w", err)
	}
	var raw struct {
		Format struct {
			Duration string `json:"duration"`
			Name     string `json:"format_name"`
		} `json:"format"`
		Streams []struct {
			Codec string `json:"codec_name"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return Probe{}, err
	}
	sec, _ := strconv.ParseFloat(raw.Format.Duration, 64)
	p := Probe{DurationMS: int(sec * 1000), Format: raw.Format.Name}
	if len(raw.Streams) > 0 {
		p.Codec = raw.Streams[0].Codec
	}
	return p, nil
}

func (e *Engine) HLS(ctx context.Context, src, destDir string, bitrate string) (int, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}
	out := filepath.Join(destDir, "index.m3u8")
	args := []string{
		"-y", "-i", src,
		"-c:a", "aac", "-b:a", bitrate, "-ac", "2", "-ar", "44100",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%.0f", stream.SegmentDurationSec),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(destDir, "seg_%d.ts"),
		out,
	}
	cmd := exec.CommandContext(ctx, e.FFmpeg, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffmpeg hls: %s: %w", stderr.String(), err)
	}
	n := 0
	entries, _ := os.ReadDir(destDir)
	for _, ent := range entries {
		if strings.HasPrefix(ent.Name(), "seg_") && strings.HasSuffix(ent.Name(), ".ts") {
			n++
		}
	}
	return n, nil
}

func (e *Engine) Peaks(ctx context.Context, src, dest string, points int) error {
	if points <= 0 {
		points = 8000
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	raw := dest + ".pcm"
	cmd := exec.CommandContext(ctx, e.FFmpeg, "-y", "-i", src, "-ac", "1", "-ar", "8000", "-f", "s16le", raw)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg pcm: %s: %w", stderr.String(), err)
	}
	defer os.Remove(raw)
	pcm, err := os.ReadFile(raw)
	if err != nil {
		return err
	}
	samples := len(pcm) / 2
	if samples == 0 {
		return fmt.Errorf("empty pcm")
	}
	step := samples / points
	if step < 1 {
		step = 1
		points = samples
	}
	peaks := make([]float64, 0, points)
	for i := 0; i < points; i++ {
		start := i * step
		end := start + step
		if end > samples {
			end = samples
		}
		var max int
		for s := start; s < end; s++ {
			v := int(int16(pcm[s*2]) | int16(pcm[s*2+1])<<8)
			if v < 0 {
				v = -v
			}
			if v > max {
				max = v
			}
		}
		peaks = append(peaks, float64(max)/32768.0)
	}
	body, err := json.Marshal(map[string]any{"version": 1, "channels": [][]float64{peaks}, "length": len(peaks), "sample_rate": 8000})
	if err != nil {
		return err
	}
	if len(body) > 200*1024 {
		logx.Warn("peaks oversized, downsampling", "bytes", len(body))
		return e.Peaks(ctx, src, dest, points/2)
	}
	return os.WriteFile(dest, body, 0o644)
}

func (e *Engine) Cover(ctx context.Context, src, dest string) error {
	_ = os.MkdirAll(filepath.Dir(dest), 0o755)
	cmd := exec.CommandContext(ctx, e.FFmpeg, "-y", "-i", src, "-an", "-vcodec", "mjpeg", "-frames:v", "1", dest)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (e *Engine) Available() bool {
	_, err := exec.LookPath(e.FFmpeg)
	if err != nil {
		if _, err2 := os.Stat(e.FFmpeg); err2 != nil {
			return false
		}
	}
	_, err = exec.LookPath(e.FFprobe)
	return err == nil
}
