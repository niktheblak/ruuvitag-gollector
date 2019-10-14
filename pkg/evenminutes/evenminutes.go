package evenminutes

import (
	"time"
)

func Next(now time.Time, interval time.Duration) time.Time {
	if interval < time.Minute {
		return now.Add(interval)
	}
	intervalInMinutes := int(interval / time.Minute)
	minute := now.Minute() + intervalInMinutes - now.Minute()%intervalInMinutes
	evenMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, 0, 0, now.Location())
	return evenMinute
}

func Until(now time.Time, interval time.Duration) time.Duration {
	ts := Next(now, interval)
	return ts.Sub(now)
}
