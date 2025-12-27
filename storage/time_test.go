package storage

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	date, err := ParseDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}
	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, date)
	}
}

func TestParseWhen(t *testing.T) {
	fallback := time.Date(2024, 1, 15, 12, 0, 0, 0, time.Local)

	// Test empty string returns fallback
	result, err := ParseWhen("", fallback)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Equal(fallback) {
		t.Errorf("Expected fallback time, got %v", result)
	}

	// Test ISO format
	result, err = ParseWhen("2024-01-15T14:30:00Z", fallback)
	if err != nil {
		t.Fatalf("Failed to parse ISO: %v", err)
	}
	expected := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test HH:MM format (for today)
	result, err = ParseWhen("14:30", fallback)
	if err != nil {
		t.Fatalf("Failed to parse HH:MM: %v", err)
	}
	expected = time.Date(2024, 1, 15, 14, 30, 0, 0, fallback.Location())
	if result.Hour() != expected.Hour() || result.Minute() != expected.Minute() {
		t.Errorf("Expected hour=%d minute=%d, got hour=%d minute=%d",
			expected.Hour(), expected.Minute(), result.Hour(), result.Minute())
	}
}

func TestToUTC(t *testing.T) {
	localTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.Local)
	utcTime := ToUTC(localTime)
	if utcTime.Location() != time.UTC {
		t.Errorf("Expected UTC timezone, got %v", utcTime.Location())
	}

	alreadyUTC := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result := ToUTC(alreadyUTC)
	if !result.Equal(alreadyUTC) {
		t.Errorf("UTC time should remain unchanged")
	}
}

func TestUTCNow(t *testing.T) {
	now := UTCNow()
	if now.Location() != time.UTC {
		t.Errorf("Expected UTC timezone, got %v", now.Location())
	}
	if now.Nanosecond() != 0 {
		t.Errorf("Expected no microseconds, got %d nanoseconds", now.Nanosecond())
	}
}

