package clock

import (
	"time"
)

// Beijing is GMT+8. All civil-date and expiry decisions must use this zone.
var Beijing = time.FixedZone("CST", 8*60*60)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func NowUTC() time.Time {
	return Now().UTC()
}

// CivilDate returns the calendar date in Asia/Shanghai.
func CivilDate(t time.Time) (year int, month time.Month, day int) {
	return t.In(Beijing).Date()
}

func FormatDisplay(t time.Time) string {
	return t.In(Beijing).Format("2006-01-02 15:04:05")
}

func ParseDisplay(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, Beijing)
}
