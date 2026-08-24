package stream

import "testing"

func TestRungForTier(t *testing.T) {
	if RungForTier("PREVIEW").Name != "128k" {
		t.Fatal("preview 128")
	}
	if RungForTier("FAN_ONLY").Name != "256k" {
		t.Fatal("fan 256")
	}
}

func TestExpectedSegments(t *testing.T) {
	if ExpectedSegments(36000, 6000) != 6 {
		t.Fatal("36s / 6s = 6")
	}
	if ExpectedSegments(30001, 6000) != 6 {
		t.Fatal("just over 30s still 6 segs")
	}
}
