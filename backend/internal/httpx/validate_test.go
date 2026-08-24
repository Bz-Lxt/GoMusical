package httpx

import "testing"

func TestEmail(t *testing.T) {
	if !Email("listener@gomusical.local") {
		t.Fatal("valid")
	}
	if Email("nope") || Email("") {
		t.Fatal("invalid accepted")
	}
}

func TestSHA256Hex(t *testing.T) {
	ok := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if !SHA256Hex(ok) {
		t.Fatal("hex")
	}
	if SHA256Hex("zz") || SHA256Hex(ok[:63]+"g") {
		t.Fatal("bad hex")
	}
}

func TestPreviewSeconds(t *testing.T) {
	if PreviewSeconds(20) || !PreviewSeconds(30) {
		t.Fatal("preview enum")
	}
}

func TestFilename(t *testing.T) {
	if Filename("../x.wav") || Filename("a/b.wav") {
		t.Fatal("escape")
	}
	if !Filename("河对岸的灯.wav") {
		t.Fatal("utf name")
	}
}
