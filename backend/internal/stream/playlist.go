package stream

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const SegmentDurationSec = 6.0
const SegmentDurationMS = 6000

// Playlist builds a VOD m3u8. It is generated at request time — never a static file.
func Playlist(baseURL string, segmentCount int, maxIndex int, targetDur float64) string {
	if segmentCount < 0 {
		segmentCount = 0
	}
	if maxIndex < 0 {
		maxIndex = -1
	}
	if maxIndex >= segmentCount {
		maxIndex = segmentCount - 1
	}
	if targetDur <= 0 {
		targetDur = SegmentDurationSec
	}
	var b bytes.Buffer
	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	b.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(targetDur+0.999)))
	b.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
	b.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	base := strings.TrimRight(baseURL, "/")
	for i := 0; i <= maxIndex; i++ {
		b.WriteString(fmt.Sprintf("#EXTINF:%.3f,\n", targetDur))
		b.WriteString(base)
		b.WriteString("/seg_")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(".ts\n")
	}
	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

func ParseSegName(name string) (int, bool) {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "seg_") || !strings.HasSuffix(name, ".ts") {
		return 0, false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(name, "seg_"), ".ts")
	if body == "" {
		return 0, false
	}
	n := 0
	for _, c := range body {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
