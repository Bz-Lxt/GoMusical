package service

import (
	"strings"
	"unicode/utf8"

	"gomusical/internal/httpx"
	"gomusical/internal/model"
)

type CommentDraft struct {
	TimestampMS int
	Body        string
}

func NormalizeComment(d CommentDraft) (CommentDraft, error) {
	d.Body = strings.TrimSpace(d.Body)
	if !httpx.CommentBody(d.Body) {
		return d, httpx.ErrBadRequest
	}
	if d.TimestampMS < 0 {
		return d, httpx.ErrCommentWindow
	}
	return d, nil
}

func CanModerate(actor *model.User, track *model.Track) bool {
	if actor == nil || track == nil {
		return false
	}
	return actor.Role == model.RoleAdmin || actor.ID == track.CreatorID
}

func VisibleTo(list []model.Comment, viewer *model.User, track *model.Track) []model.Comment {
	out := []model.Comment{}
	mod := CanModerate(viewer, track)
	for _, c := range list {
		if c.Hidden && !mod {
			continue
		}
		out = append(out, c)
	}
	return out
}

func SortKey(c model.Comment) int {
	if c.Pinned {
		return c.TimestampMS - 1_000_000_000
	}
	return c.TimestampMS
}

func PreviewSnippet(body string, n int) string {
	if n <= 0 {
		n = 24
	}
	if utf8.RuneCountInString(body) <= n {
		return body
	}
	r := []rune(body)
	return string(r[:n]) + "…"
}
