package transcode

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeRe = regexp.MustCompile(`time=(\d+):(\d+):(\d+\.?\d*)`)

// ParseFFmpegTime extracts the last time=HH:MM:SS.xx from ffmpeg stderr.
func ParseFFmpegTime(stderr string) (time.Duration, bool) {
	matches := timeRe.FindAllStringSubmatch(stderr, -1)
	if len(matches) == 0 {
		return 0, false
	}
	m := matches[len(matches)-1]
	h, _ := strconv.Atoi(m[1])
	min, _ := strconv.Atoi(m[2])
	sec, _ := strconv.ParseFloat(m[3], 64)
	d := time.Duration(h)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec*float64(time.Second))
	return d, true
}

func ProgressPercent(done, total time.Duration) int {
	if total <= 0 {
		return 0
	}
	p := int(float64(done) / float64(total) * 100)
	if p < 0 {
		return 0
	}
	if p > 99 {
		return 99
	}
	return p
}

func ClassifyFFmpegError(stderr string) string {
	s := strings.ToLower(stderr)
	switch {
	case strings.Contains(s, "no such file"):
		return "source_missing"
	case strings.Contains(s, "invalid data found"):
		return "invalid_media"
	case strings.Contains(s, "permission denied"):
		return "permission"
	case strings.Contains(s, "killed") || strings.Contains(s, "signal"):
		return "killed"
	default:
		if len(stderr) > 240 {
			return "ffmpeg: " + stderr[len(stderr)-240:]
		}
		return "ffmpeg: " + stderr
	}
}
