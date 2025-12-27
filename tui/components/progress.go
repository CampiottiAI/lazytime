package components

import (
	"fmt"
	"lazytime/storage"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderProgressBar renders a progress bar for goal tracking.
func RenderProgressBar(current, target time.Duration, label string, width int, progressStyle lipgloss.Style, formatDuration func(time.Duration) string) string {
	if target <= 0 {
		return label + ": N/A"
	}

	percent := float64(current) / float64(target)
	if percent > 1.0 {
		percent = 1.0
	}

	barWidth := width - 20 // Leave space for text
	if barWidth < 10 {
		barWidth = 10
	}

	filled := int(float64(barWidth) * percent)
	empty := barWidth - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	styledBar := progressStyle.Render(bar)

	percentText := formatDuration(current) + " / " + formatDuration(target)
	percentNum := int(percent * 100)
	if percentNum > 100 {
		percentNum = 100
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		label+":",
		styledBar,
		fmt.Sprintf("%d%%", percentNum),
		"("+percentText+")",
	)
}

// RenderGoalProgress renders progress bars for daily and weekly goals.
func RenderGoalProgress(entries []storage.Entry, now time.Time, targetToday, targetWeek time.Duration, width int, clampDuration func(storage.Entry, time.Time, time.Time, time.Time) time.Duration, getProgressStyle func(time.Duration, time.Duration) lipgloss.Style, formatDuration func(time.Duration) string) string {
	tz := now.Location()
	today := now

	// Calculate today's total
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
	todayEnd := todayStart.AddDate(0, 0, 1)
	todayStartUTC := storage.ToUTC(todayStart)
	todayEndUTC := storage.ToUTC(todayEnd)

	var todayTotal time.Duration
	for _, entry := range entries {
		todayTotal += clampDuration(entry, todayStartUTC, todayEndUTC, now)
	}

	// Calculate week's total
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday-- // Monday = 0
	weekStart := today.AddDate(0, 0, -weekday)
	weekStartLocal := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, tz)
	weekEndLocal := weekStartLocal.AddDate(0, 0, 7)
	weekStartUTC := storage.ToUTC(weekStartLocal)
	weekEndUTC := storage.ToUTC(weekEndLocal)

	var weekTotal time.Duration
	for _, entry := range entries {
		weekTotal += clampDuration(entry, weekStartUTC, weekEndUTC, now)
	}

	todayBar := RenderProgressBar(todayTotal, targetToday, "Today", width, getProgressStyle(todayTotal, targetToday), formatDuration)
	weekBar := RenderProgressBar(weekTotal, targetWeek, "Week", width, getProgressStyle(weekTotal, targetWeek), formatDuration)

	return lipgloss.JoinVertical(lipgloss.Left, todayBar, weekBar)
}

