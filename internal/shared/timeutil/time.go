package timeutil

import "time"

const (
	HourSeconds = 3600
	DaySeconds  = 24 * HourSeconds
)

func HourStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
}

func UnixToTimePtr(from *int32) *time.Time {
	if from == nil {
		return nil
	}

	to := time.Unix(int64(*from), 0).UTC()
	return &to
}
