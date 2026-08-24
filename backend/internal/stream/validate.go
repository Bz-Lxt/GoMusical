package stream

import (
	"fmt"
	"unicode"
)

func AssertSafeToken(tok string) error {
	if tok == "" || len(tok) > 2048 {
		return fmt.Errorf("token length")
	}
	for _, r := range tok {
		if r > unicode.MaxASCII {
			return fmt.Errorf("token charset")
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '~') {
			return fmt.Errorf("token charset")
		}
	}
	return nil
}

func SegmentStartMS(index int) int {
	if index < 0 {
		return 0
	}
	return index * SegmentDurationMS
}

func InWindow(index, untilMS int) bool {
	return SegmentStartMS(index) < untilMS
}
