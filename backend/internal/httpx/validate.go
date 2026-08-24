package httpx

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

func Email(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 190 {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

func DisplayName(s string) bool {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return n >= 2 && n <= 40
}

func CommentBody(s string) bool {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return n >= 1 && n <= 500
}

func PreviewSeconds(n int) bool {
	return n == 15 || n == 30 || n == 60
}

func PriceCents(n int) bool {
	return n >= 0 && n <= 1_000_000
}

func SHA256Hex(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func Filename(s string) bool {
	if s == "" || len(s) > 180 || strings.Contains(s, "..") || strings.ContainsAny(s, `/\`) {
		return false
	}
	return true
}
