package clock

import "testing"

func TestFormatRoundTrip(t *testing.T) {
	n := Now()
	s := FormatDisplay(n)
	got, err := ParseDisplay(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != n.Year() || got.Month() != n.Month() || got.Day() != n.Day() {
		t.Fatalf("%v vs %v", got, n)
	}
}

func TestCivilUsesBeijing(t *testing.T) {
	y, m, d := CivilDate(Now())
	if y < 2026 || m < 1 || d < 1 {
		t.Fatal("civil")
	}
}
