package api

import (
	"fmt"
	"time"
)

func SnoozeUntil(preset string, now time.Time) (time.Time, error) {
	loc := now.Location()
	y, m, d := now.Date()
	switch preset {
	case "hour":
		return now.Add(time.Hour), nil
	case "tomorrow":
		return time.Date(y, m, d, 9, 0, 0, 0, loc).AddDate(0, 0, 1), nil
	case "week":
		return time.Date(y, m, d, 9, 0, 0, 0, loc).AddDate(0, 0, 7), nil
	default:
		return time.Time{}, fmt.Errorf("unknown snooze preset %q", preset)
	}
}