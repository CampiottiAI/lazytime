package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"pytimelog/storage"
)

// FormatDuration formats a duration as "XhYYm".
func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	return fmt.Sprintf("%dh%02dm", hours, minutes)
}

// ClampDuration calculates the overlap duration of an entry within a time range.
func ClampDuration(entry storage.Entry, start, end, now time.Time) time.Duration {
	entryEnd := now
	if entry.End != nil {
		entryEnd = *entry.End
	}

	latestStart := entry.Start
	if start.After(latestStart) {
		latestStart = start
	}

	earliestEnd := entryEnd
	if end.Before(earliestEnd) {
		earliestEnd = end
	}

	if earliestEnd.Before(latestStart) || earliestEnd.Equal(latestStart) {
		return 0
	}

	return earliestEnd.Sub(latestStart)
}

// Summarize aggregates entries by tag within a time range.
// Returns total duration and a map of tag -> duration.
func Summarize(entries []storage.Entry, start, end, now time.Time) (time.Duration, map[string]time.Duration) {
	tagTotals := make(map[string]time.Duration)
	var total time.Duration

	for _, entry := range entries {
		chunk := ClampDuration(entry, start, end, now)
		if chunk <= 0 {
			continue
		}
		total += chunk

		tags := entry.Tags()
		if len(tags) == 0 {
			tags = []string{"(untagged)"}
		}

		for _, tag := range tags {
			tagTotals[tag] += chunk
		}
	}

	return total, tagTotals
}

// CommandStart starts a new active entry.
func CommandStart(text string, atTime string) error {
	entries, err := storage.ReadEntries("")
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	openIdx := storage.FindOpen(entries)
	if openIdx != -1 {
		return fmt.Errorf("there is already an active entry. Stop it before starting another")
	}

	now := storage.LocalNow()
	when, err := storage.ParseWhen(atTime, now)
	if err != nil {
		return err
	}

	whenUTC := storage.ToUTC(when)
	newEntry := storage.Entry{
		Start: whenUTC,
		End:   nil,
		Text:  text,
	}

	if err := storage.AppendEntry(newEntry, ""); err != nil {
		return fmt.Errorf("failed to append entry: %w", err)
	}

	localWhen := whenUTC.In(now.Location())
	fmt.Printf("Started: %s @ %s\n", text, localWhen.Format("2006-01-02 15:04"))
	return nil
}

// CommandStop stops the active entry.
func CommandStop(atTime string) error {
	entries, err := storage.ReadEntries("")
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	openIdx := storage.FindOpen(entries)
	if openIdx == -1 {
		return fmt.Errorf("no active entry to stop")
	}

	now := storage.LocalNow()
	when, err := storage.ParseWhen(atTime, now)
	if err != nil {
		return err
	}

	whenUTC := storage.ToUTC(when)
	openEntry := entries[openIdx]

	if whenUTC.Before(openEntry.Start) || whenUTC.Equal(openEntry.Start) {
		return fmt.Errorf("stop time must be after the start time")
	}

	updated := storage.Entry{
		Start: openEntry.Start,
		End:   &whenUTC,
		Text:  openEntry.Text,
	}

	entries[openIdx] = updated
	if err := storage.WriteEntries(entries, ""); err != nil {
		return fmt.Errorf("failed to write entries: %w", err)
	}

	elapsed := updated.Duration(whenUTC)
	fmt.Printf("Stopped '%s' after %s.\n", updated.Text, FormatDuration(elapsed))
	return nil
}

// CommandAdd adds a completed entry retroactively.
func CommandAdd(start, end, text string) error {
	entries, err := storage.ReadEntries("")
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	now := storage.LocalNow()
	startTime, err := storage.ParseWhen(start, now)
	if err != nil {
		return err
	}
	endTime, err := storage.ParseWhen(end, now)
	if err != nil {
		return err
	}

	startUTC := storage.ToUTC(startTime)
	endUTC := storage.ToUTC(endTime)

	if endUTC.Before(startUTC) || endUTC.Equal(startUTC) {
		return fmt.Errorf("end time must be after start time")
	}

	newEntry := storage.Entry{
		Start: startUTC,
		End:   &endUTC,
		Text:  text,
	}

	overlapEntry, overlapDuration, hasOverlap := storage.CheckOverlap(entries, newEntry, endUTC)
	if hasOverlap {
		otherLocal := overlapEntry.Start.In(now.Location())
		return fmt.Errorf(
			"new entry overlaps with existing entry starting at %s for %s",
			otherLocal.Format("2006-01-02 15:04"),
			FormatDuration(overlapDuration),
		)
	}

	if err := storage.AppendEntry(newEntry, ""); err != nil {
		return fmt.Errorf("failed to append entry: %w", err)
	}

	startLocal := startUTC.In(now.Location())
	endLocal := endUTC.In(now.Location())
	fmt.Printf(
		"Added %s entry %s -> %s : %s\n",
		FormatDuration(newEntry.Duration(endUTC)),
		startLocal.Format("2006-01-02 15:04"),
		endLocal.Format("15:04"),
		text,
	)
	return nil
}

// CommandStatus shows the current active entry.
func CommandStatus() error {
	entries, err := storage.ReadEntries("")
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	openIdx := storage.FindOpen(entries)
	if openIdx == -1 {
		fmt.Println("No active entry.")
		return nil
	}

	entry := entries[openIdx]
	now := storage.UTCNow()
	elapsed := entry.Duration(now)
	localStart := entry.Start.In(time.Local)

	fmt.Printf(
		"Active: %s (since %s, %s elapsed)\n",
		entry.Text,
		localStart.Format("15:04"),
		FormatDuration(elapsed),
	)
	return nil
}

// CommandReport generates a report of logged time by tag for a date range.
func CommandReport(fromDate, toDate string, week, lastWeek bool) error {
	entries, err := storage.ReadEntries("")
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	now := storage.LocalNow()
	tz := now.Location()
	today := now

	if week && lastWeek {
		return fmt.Errorf("choose only one of --week or --last-week")
	}
	if (week || lastWeek) && (fromDate != "" || toDate != "") {
		return fmt.Errorf("cannot combine --week/--last-week with --from/--to")
	}

	var from, to time.Time
	if week {
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		weekday-- // Monday = 0
		from = today.AddDate(0, 0, -weekday)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, tz)
		to = from.AddDate(0, 0, 6)
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, tz)
	} else if lastWeek {
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		weekday--
		thisWeekStart := today.AddDate(0, 0, -weekday)
		from = thisWeekStart.AddDate(0, 0, -7)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, tz)
		to = from.AddDate(0, 0, 6)
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, tz)
	} else {
		if fromDate == "" {
			from = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
		} else {
			parsed, err := storage.ParseDate(fromDate)
			if err != nil {
				return fmt.Errorf("invalid from date: %w", err)
			}
			from = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, tz)
		}

		if toDate == "" {
			to = from
		} else {
			parsed, err := storage.ParseDate(toDate)
			if err != nil {
				return fmt.Errorf("invalid to date: %w", err)
			}
			to = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, tz)
		}
	}

	if to.Before(from) {
		return fmt.Errorf("report end date cannot be before start date")
	}

	startUTC := from.UTC()
	endUTC := to.UTC()
	nowUTC := storage.UTCNow()

	total, tagTotals := Summarize(entries, startUTC, endUTC, nowUTC)

	if total == 0 {
		fmt.Println("No entries in the selected range.")
		return nil
	}

	fromDateStr := from.Format("2006-01-02")
	toDateStr := to.Format("2006-01-02")
	fmt.Printf("Report %s to %s\n", fromDateStr, toDateStr)

	// Sort tags case-insensitively but preserve original spelling
	type tagItem struct {
		tag      string
		duration time.Duration
	}
	var sortedTags []tagItem
	for tag, duration := range tagTotals {
		sortedTags = append(sortedTags, tagItem{tag: tag, duration: duration})
	}
	sort.Slice(sortedTags, func(i, j int) bool {
		return strings.ToLower(sortedTags[i].tag) < strings.ToLower(sortedTags[j].tag)
	})

	for _, item := range sortedTags {
		fmt.Printf("- %s: %s\n", item.tag, FormatDuration(item.duration))
	}
	fmt.Printf("Total: %s\n", FormatDuration(total))

	return nil
}

// CommandTUI launches the terminal UI.
func CommandTUI() error {
	// Import tui package and call LaunchTUI
	// We'll handle this in main.go to avoid circular dependencies
	return fmt.Errorf("TUI should be called from main")
}

// RunCLI parses command-line arguments and executes the appropriate command.
func RunCLI(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	remaining := args[1:]

	switch command {
	case "start":
		if len(remaining) == 0 {
			return fmt.Errorf("start command requires text argument")
		}
		text := remaining[0]
		var atTime string
		if len(remaining) > 1 && remaining[1] == "--at" {
			if len(remaining) < 3 {
				return fmt.Errorf("--at requires a time value")
			}
			atTime = remaining[2]
			remaining = remaining[3:]
		}
		// Handle text that might have been split
		if len(remaining) > 0 {
			text = strings.Join(append([]string{text}, remaining...), " ")
		}
		return CommandStart(text, atTime)

	case "stop":
		var atTime string
		if len(remaining) > 0 && remaining[0] == "--at" {
			if len(remaining) < 2 {
				return fmt.Errorf("--at requires a time value")
			}
			atTime = remaining[1]
		}
		return CommandStop(atTime)

	case "add":
		var start, end, text string
		hasStart := false
		hasEnd := false
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == "--start" {
				if i+1 >= len(remaining) {
					return fmt.Errorf("--start requires a time value")
				}
				start = remaining[i+1]
				hasStart = true
				i++
			} else if remaining[i] == "--end" {
				if i+1 >= len(remaining) {
					return fmt.Errorf("--end requires a time value")
				}
				end = remaining[i+1]
				hasEnd = true
				i++
			} else {
				text = remaining[i]
			}
		}
		if !hasStart || !hasEnd {
			return fmt.Errorf("add command requires --start and --end")
		}
		if text == "" {
			return fmt.Errorf("add command requires text argument")
		}
		return CommandAdd(start, end, text)

	case "status":
		return CommandStatus()

	case "report":
		var fromDate, toDate string
		week := false
		lastWeek := false
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == "--from" {
				if i+1 >= len(remaining) {
					return fmt.Errorf("--from requires a date value")
				}
				fromDate = remaining[i+1]
				i++
			} else if remaining[i] == "--to" {
				if i+1 >= len(remaining) {
					return fmt.Errorf("--to requires a date value")
				}
				toDate = remaining[i+1]
				i++
			} else if remaining[i] == "--week" {
				week = true
			} else if remaining[i] == "--last-week" {
				lastWeek = true
			}
		}
		return CommandReport(fromDate, toDate, week, lastWeek)

	case "tui":
		return CommandTUI()

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

