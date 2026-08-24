package stream

// BitrateLadder is the locked streaming ladder (C-4): AAC only, never FLAC/WAV.
type Rung struct {
	Name    string
	Bitrate string
	ForTier []string
}

var DefaultLadder = []Rung{
	{Name: "128k", Bitrate: "128k", ForTier: []string{"PREVIEW"}},
	{Name: "256k", Bitrate: "256k", ForTier: []string{"PAID_DOWNLOAD", "FAN_ONLY"}},
}

func RungForTier(tier string) Rung {
	for _, r := range DefaultLadder {
		for _, t := range r.ForTier {
			if t == tier {
				return r
			}
		}
	}
	return DefaultLadder[0]
}

func SegmentFile(dir, rung string, index int) string {
	return dir + "/" + rung + "/seg_" + itoa(index) + ".ts"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func ExpectedSegments(durationMS, segMS int) int {
	if durationMS <= 0 || segMS <= 0 {
		return 0
	}
	n := durationMS / segMS
	if durationMS%segMS != 0 {
		n++
	}
	return n
}
