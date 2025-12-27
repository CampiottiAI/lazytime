package storage

import (
	"testing"
	"time"
)

func TestFormatAndParseRoundTrip(t *testing.T) {
	entry := Entry{
		Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		End:   func() *time.Time { t := time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC); return &t }(),
		Text:  "Write docs #project",
	}

	raw := FormatEntry(entry)
	parsed, err := ParseEntry(raw)
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if !parsed.Start.Equal(entry.Start) {
		t.Errorf("Start time mismatch: got %v, want %v", parsed.Start, entry.Start)
	}
	if parsed.End == nil || entry.End == nil {
		if parsed.End != entry.End {
			t.Errorf("End time nil mismatch: got %v, want %v", parsed.End, entry.End)
		}
	} else if !parsed.End.Equal(*entry.End) {
		t.Errorf("End time mismatch: got %v, want %v", parsed.End, entry.End)
	}
	if parsed.Text != entry.Text {
		t.Errorf("Text mismatch: got %q, want %q", parsed.Text, entry.Text)
	}
}

func TestOverlapDetection(t *testing.T) {
	first := Entry{
		Start: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		End:   func() *time.Time { t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC); return &t }(),
		Text:  "Morning work",
	}

	overlapping := Entry{
		Start: time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC),
		End:   func() *time.Time { t := time.Date(2024, 1, 1, 9, 45, 0, 0, time.UTC); return &t }(),
		Text:  "Conflicts",
	}

	entries := []Entry{first}
	overlapEntry, overlapDuration, hasOverlap := CheckOverlap(entries, overlapping, *first.End)

	if !hasOverlap {
		t.Error("Expected overlap to be detected")
	}
	if overlapEntry.Text != first.Text {
		t.Errorf("Expected overlapping entry to be %q, got %q", first.Text, overlapEntry.Text)
	}
	expectedDuration := 15 * time.Minute
	if overlapDuration != expectedDuration {
		t.Errorf("Expected overlap duration %v, got %v", expectedDuration, overlapDuration)
	}
}

func TestParseTimeOfDay(t *testing.T) {
	hour, minute, err := ParseTimeOfDay("09:05")
	if err != nil {
		t.Fatalf("Failed to parse time: %v", err)
	}
	if hour != 9 || minute != 5 {
		t.Errorf("Expected hour=9, minute=5, got hour=%d, minute=%d", hour, minute)
	}

	_, _, err = ParseTimeOfDay("99:99")
	if err == nil {
		t.Error("Expected error for invalid time 99:99")
	}
}

func TestEntryTags(t *testing.T) {
	entry := Entry{
		Text: "Write docs #project #writing",
	}
	tags := entry.Tags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "project" || tags[1] != "writing" {
		t.Errorf("Expected tags [project writing], got %v", tags)
	}

	entryNoTags := Entry{
		Text: "Write docs",
	}
	tags = entryNoTags.Tags()
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}

func TestEntryDuration(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	entry := Entry{
		Start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   func() *time.Time { t := time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC); return &t }(),
		Text:  "Test",
	}

	duration := entry.Duration(now)
	expected := 90 * time.Minute
	if duration != expected {
		t.Errorf("Expected duration %v, got %v", expected, duration)
	}

	openEntry := Entry{
		Start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   nil,
		Text:  "Open",
	}

	duration = openEntry.Duration(now)
	expected = 2 * time.Hour
	if duration != expected {
		t.Errorf("Expected duration %v for open entry, got %v", expected, duration)
	}
}

func TestFindOpen(t *testing.T) {
	entries := []Entry{
		{
			Start: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			End:   func() *time.Time { t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC); return &t }(),
			Text:  "Closed",
		},
		{
			Start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			End:   nil,
			Text:  "Open",
		},
	}

	idx := FindOpen(entries)
	if idx != 1 {
		t.Errorf("Expected open entry at index 1, got %d", idx)
	}

	allClosed := []Entry{
		{
			Start: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			End:   func() *time.Time { t := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC); return &t }(),
			Text:  "Closed",
		},
	}

	idx = FindOpen(allClosed)
	if idx != -1 {
		t.Errorf("Expected no open entry (-1), got %d", idx)
	}
}

