package schedule

import "time"

var validSchedules = map[string]time.Duration{
	"hourly":  time.Hour,
	"daily":   time.Hour * 24,
	"weekly":  time.Hour * 24 * 7,
	"monthly": time.Hour * 24 * 7 * 4,
}

func IsValid(schedule string) bool {
	_, found := validSchedules[schedule]
	return found
}

func ShouldProcess(schedule string, lastProcessed *time.Time) bool {
	return time.Since(*lastProcessed) >= validSchedules[schedule]
}
