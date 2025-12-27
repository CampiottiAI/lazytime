package storage

import (
	"fmt"
	"regexp"
	"time"
)

// UTCNow returns current UTC time with seconds precision (no microseconds).
func UTCNow() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

// ParseTimeOfDay parses a time string in HH:MM format.
// Returns hour and minute, or an error if invalid.
func ParseTimeOfDay(value string) (hour, minute int, err error) {
	re := regexp.MustCompile(`^(?P<hour>\d{1,2}):(?P<minute>\d{2})$`)
	matches := re.FindStringSubmatch(value)
	if matches == nil {
		return 0, 0, fmt.Errorf("invalid time format: %s", value)
	}

	var h, m int
	fmt.Sscanf(matches[1], "%d", &h)
	fmt.Sscanf(matches[2], "%d", &m)

	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid time value: %s", value)
	}

	return h, m, nil
}

// ParseWhen parses a time string that can be either:
// - An ISO 8601 datetime string (with optional timezone)
// - An HH:MM time for today in local timezone
// If value is empty, returns fallback.
func ParseWhen(value string, fallback time.Time) (time.Time, error) {
	if value == "" {
		return fallback, nil
	}

	// Try ISO format first
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
		// No timezone, use fallback's timezone
		loc := fallback.Location()
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc), nil
	}

	// Try HH:MM for today
	hour, minute, err := ParseTimeOfDay(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse time: %s", value)
	}

	today := fallback.Local()
	return time.Date(today.Year(), today.Month(), today.Day(), hour, minute, 0, 0, today.Location()), nil
}

// ToUTC converts a time to UTC, handling nil timezone by assuming UTC.
func ToUTC(value time.Time) time.Time {
	if value.Location() == time.UTC {
		return value
	}
	if value.Location() == time.Local {
		// If it's local, convert it
		return value.UTC()
	}
	// If timezone is nil or unknown, assume UTC
	if value.Location().String() == "" {
		return value.UTC()
	}
	return value.UTC()
}

// LocalNow returns current local time with seconds precision (no microseconds).
func LocalNow() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
}

