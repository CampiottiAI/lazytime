package components

import (
	"lazytime/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// removeTags removes #tag patterns from text.
func removeTags(text string) string {
	words := strings.Fields(text)
	var cleaned []string
	for _, word := range words {
		if !strings.HasPrefix(word, "#") {
			cleaned = append(cleaned, word)
		}
	}
	return strings.Join(cleaned, " ")
}

// RenderHero renders the hero section with large timer and current task info.
func RenderHero(entries []storage.Entry, now time.Time, width int, borderIdle, borderRunning, styleIdle, heroTimerStyle, heroTaskStyle, heroTagStyle lipgloss.Style, getTagColor func(string) lipgloss.Color, formatDuration, formatDurationShort, formatDurationFull func(time.Duration) string, clampDuration func(storage.Entry, time.Time, time.Time, time.Time) time.Duration) string {
	idx := storage.FindOpen(entries)

	var lines []string

	if idx == -1 {
		// No active task - show idle state
		// Calculate idle duration: time since last entry ended (or 00:00:00 if no entries today)
		tz := now.Location()
		today := now
		todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
		todayEnd := todayStart.AddDate(0, 0, 1)
		todayStartUTC := storage.ToUTC(todayStart)
		todayEndUTC := storage.ToUTC(todayEnd)

		// Filter entries for today
		var todayEntries []storage.Entry
		for _, entry := range entries {
			if clampDuration(entry, todayStartUTC, todayEndUTC, now) > 0 {
				todayEntries = append(todayEntries, entry)
			}
		}

		var idleDuration time.Duration
		if len(todayEntries) == 0 {
			// No entries today - show 00:00:00
			idleDuration = 0
		} else {
			// Find the most recent entry's end time
			var lastEnd time.Time
			for _, entry := range todayEntries {
				entryEnd := now
				if entry.End != nil {
					entryEnd = *entry.End
				}
				if entryEnd.After(lastEnd) {
					lastEnd = entryEnd
				}
			}
			// Calculate idle duration from last entry end to now
			idleDuration = now.Sub(lastEnd)
			if idleDuration < 0 {
				idleDuration = 0
			}
		}

		// Format idle duration as HH:MM:SS
		idleText := formatDurationFull(idleDuration)
		lines = append(lines, lipgloss.Place(width-4, 1, lipgloss.Center, lipgloss.Center, styleIdle.Render("IDLE "+idleText)))
	} else {
		// Active task - show compact status with elapsed time and task description
		entry := entries[idx]
		elapsed := entry.Duration(now)

		// Format elapsed time in big bold green (always HH:MM:SS)
		timerText := formatDurationFull(elapsed)
		styledTimer := heroTimerStyle.Render(timerText)

		// Get task description without tags
		taskDescription := removeTags(entry.Text)
		styledTask := heroTaskStyle.Render(taskDescription)

		// Create horizontal layout: [ELAPSED TIME] [TASK DESCRIPTION]
		// Account for border padding (2 chars on each side = 4 total)
		availableWidth := width - 4
		timerWidth := lipgloss.Width(styledTimer)
		taskWidth := lipgloss.Width(styledTask)
		spacing := 2 // Space between timer and task

		// If content fits, use simple layout
		if timerWidth+spacing+taskWidth <= availableWidth {
			content := styledTimer + strings.Repeat(" ", spacing) + styledTask
			lines = append(lines, content)
		} else {
			// If task is too long, truncate it
			maxTaskWidth := availableWidth - timerWidth - spacing
			if maxTaskWidth > 0 {
				// Truncate task description to fit
				truncatedTask := lipgloss.Place(maxTaskWidth, 1, lipgloss.Left, lipgloss.Top, styledTask)
				content := styledTimer + strings.Repeat(" ", spacing) + truncatedTask
				lines = append(lines, content)
			} else {
				// If even timer doesn't fit, just show timer
				lines = append(lines, styledTimer)
			}
		}
	}

	// Determine border style based on state
	var borderStyle lipgloss.Style
	if idx == -1 {
		borderStyle = borderIdle
	} else {
		borderStyle = borderRunning
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return borderStyle.Width(width).Render(content)
}
