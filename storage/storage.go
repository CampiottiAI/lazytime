package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const LogEnvVar = "LAZYTIME_PATH"

// Entry represents a time log entry.
type Entry struct {
	Start time.Time
	End   *time.Time // nil for open entries
	Text  string
}

// Duration returns the duration of the entry.
// If End is nil, uses the provided now time (or UTCNow() if now is zero).
func (e Entry) Duration(now time.Time) time.Duration {
	if now.IsZero() {
		now = UTCNow()
	}
	end := now
	if e.End != nil {
		end = *e.End
	}
	return end.Sub(e.Start)
}

// Tags extracts all #tag patterns from the entry text.
// Returns an empty slice if no tags are found.
func (e Entry) Tags() []string {
	var tags []string
	words := strings.Fields(e.Text)
	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			tags = append(tags, word[1:])
		}
	}
	return tags
}

// DefaultLogPath returns the log file path from environment variable
// or defaults to ~/.lazytime/log.txt.
func DefaultLogPath() string {
	envValue := os.Getenv(LogEnvVar)
	if envValue != "" {
		return filepath.Clean(envValue)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		return ".lazytime/log.txt"
	}
	return filepath.Join(home, ".lazytime", "log.txt")
}

// ensureAware ensures a datetime has timezone info, defaulting to UTC.
func ensureAware(value time.Time) time.Time {
	if value.Location() == time.UTC {
		return value
	}
	if value.Location() == time.Local {
		return value
	}
	// If timezone is nil or unknown, assume UTC
	if value.Location().String() == "" {
		return value.UTC()
	}
	return value
}

// FormatEntry formats an entry as a line in the log file.
// Format: ISO_START ISO_END|text or ISO_START -|text for open entries.
func FormatEntry(entry Entry) string {
	start := ensureAware(entry.Start).UTC()
	startStr := start.Format(time.RFC3339)

	var endStr string
	if entry.End == nil {
		endStr = "-"
	} else {
		end := ensureAware(*entry.End).UTC()
		endStr = end.Format(time.RFC3339)
	}

	return fmt.Sprintf("%s %s|%s", startStr, endStr, strings.TrimSpace(entry.Text))
}

// ParseEntry parses a single line from the log file.
func ParseEntry(raw string) (Entry, error) {
	if !strings.Contains(raw, "|") {
		return Entry{}, fmt.Errorf("entry must contain '|' separator")
	}

	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return Entry{}, fmt.Errorf("entry must contain '|' separator")
	}

	timesPart := strings.TrimSpace(parts[0])
	text := strings.TrimSpace(parts[1])

	times := strings.Fields(timesPart)
	if len(times) != 2 {
		return Entry{}, fmt.Errorf("entry must have start and end column")
	}

	startRaw := times[0]
	endRaw := times[1]

	start, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		return Entry{}, fmt.Errorf("invalid start time: %w", err)
	}
	start = ensureAware(start)

	var end *time.Time
	if endRaw == "-" {
		end = nil
	} else {
		endTime, err := time.Parse(time.RFC3339, endRaw)
		if err != nil {
			return Entry{}, fmt.Errorf("invalid end time: %w", err)
		}
		endTime = ensureAware(endTime)
		end = &endTime
	}

	return Entry{
		Start: start,
		End:   end,
		Text:  text,
	}, nil
}

// ReadEntries reads all entries from the log file.
// Skips empty lines and lines starting with #.
func ReadEntries(path string) ([]Entry, error) {
	if path == "" {
		path = DefaultLogPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Read file if it exists
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	var entries []Entry
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}

		entry, err := ParseEntry(stripped)
		if err != nil {
			// Skip malformed entries but continue reading
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// WriteEntries writes all entries to the log file.
func WriteEntries(entries []Entry, path string) error {
	if path == "" {
		path = DefaultLogPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	var lines []string
	for _, entry := range entries {
		lines = append(lines, FormatEntry(entry))
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// AppendEntry appends a single entry to the log file.
func AppendEntry(entry Entry, path string) error {
	if path == "" {
		path = DefaultLogPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	line := FormatEntry(entry) + "\n"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(line)
	return err
}

// FindOpen returns the index of the first open entry (End == nil) from the end.
// Returns -1 if no open entry is found.
func FindOpen(entries []Entry) int {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].End == nil {
			return i
		}
	}
	return -1
}

// CheckOverlap checks if a candidate entry overlaps with any existing entry.
// Returns the overlapping entry, overlap duration, and true if overlap found.
func CheckOverlap(entries []Entry, candidate Entry, now time.Time) (Entry, time.Duration, bool) {
	if now.IsZero() {
		now = UTCNow()
	}

	candidateEnd := now
	if candidate.End != nil {
		candidateEnd = *candidate.End
	}

	for _, existing := range entries {
		existingEnd := now
		if existing.End != nil {
			existingEnd = *existing.End
		}

		// Check if intervals overlap
		if candidate.Start.Before(existingEnd) && candidateEnd.After(existing.Start) {
			overlapStart := candidate.Start
			if existing.Start.After(overlapStart) {
				overlapStart = existing.Start
			}
			overlapEnd := candidateEnd
			if existingEnd.Before(overlapEnd) {
				overlapEnd = existingEnd
			}
			overlapDuration := overlapEnd.Sub(overlapStart)
			if overlapDuration > 0 {
				return existing, overlapDuration, true
			}
		}
	}

	return Entry{}, 0, false
}

