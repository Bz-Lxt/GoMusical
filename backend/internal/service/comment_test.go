package service

import (
	"testing"

	"gomusical/internal/model"
)

func TestNormalizeComment(t *testing.T) {
	d, err := NormalizeComment(CommentDraft{TimestampMS: 10, Body: "  hi  "})
	if err != nil || d.Body != "hi" {
		t.Fatal(d, err)
	}
	if _, err := NormalizeComment(CommentDraft{TimestampMS: -1, Body: "x"}); err == nil {
		t.Fatal("neg ts")
	}
}

func TestVisibleHidden(t *testing.T) {
	tr := &model.Track{CreatorID: "c1"}
	list := []model.Comment{{ID: "1", Hidden: true}, {ID: "2"}}
	v := VisibleTo(list, &model.User{ID: "x", Role: model.RoleListener}, tr)
	if len(v) != 1 {
		t.Fatal(len(v))
	}
	v = VisibleTo(list, &model.User{ID: "c1", Role: model.RoleCreator}, tr)
	if len(v) != 2 {
		t.Fatal("creator sees hidden")
	}
}

func TestSnippet(t *testing.T) {
	if PreviewSnippet("短", 8) != "短" {
		t.Fatal("short")
	}
}
