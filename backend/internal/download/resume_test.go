package download

import "testing"

func TestResumeComplete(t *testing.T) {
	p := DefaultResume()
	if p.Completed(94, 100) {
		t.Fatal("94% not complete")
	}
	if !p.Completed(95, 100) {
		t.Fatal("95% complete")
	}
}

func TestAllowResumeAfterMaxUses(t *testing.T) {
	p := DefaultResume()
	if !p.AllowAnotherHit(3, 3, 10, 100) {
		t.Fatal("in-progress resume must be allowed even if uses==max")
	}
	if p.AllowAnotherHit(3, 3, 96, 100) {
		t.Fatal("finished ticket must stop")
	}
}

func TestMergeBytes(t *testing.T) {
	if MergeBytes(90, 20, 100) != 100 {
		t.Fatal("cap at size")
	}
}
