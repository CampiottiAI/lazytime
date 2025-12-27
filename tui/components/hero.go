package components

import (
	"lazytime/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderHero renders the hero section with large timer and current task info.
func RenderHero(entries []storage.Entry, now time.Time, width int, borderIdle, borderRunning, styleIdle, heroTimerStyle, heroTaskStyle, heroTagStyle lipgloss.Style, getTagColor func(string) lipgloss.Color, formatDuration, formatDurationShort func(time.Duration) string) string {
	idx := storage.FindOpen(entries)
	tz := now.Location()

	var lines []string

	if idx == -1 {
		// No active task - show idle state
		lines = append(lines, borderIdle.Render(""))
		lines = append(lines, "")
		lines = append(lines, lipgloss.Place(width-4, 1, lipgloss.Center, lipgloss.Center, styleIdle.Render("IDLE")))
		lines = append(lines, "")

		// Find last closed entry
		var closed *storage.Entry
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i].End != nil {
				closed = &entries[i]
				break
			}
		}
		if closed != nil {
			lines = append(lines, "")
			lines = append(lines, heroTaskStyle.Render("Last: "+closed.Text))
			lines = append(lines, heroTagStyle.Render("Ended: "+closed.End.In(tz).Format("2006-01-02 15:04")))
			lines = append(lines, heroTagStyle.Render("Length: "+formatDurationShort(closed.Duration(*closed.End))))
		}
	} else {
		// Active task - show large timer
		entry := entries[idx]
		startLocal := entry.Start.In(tz)
		elapsed := entry.Duration(now)

		// Large timer display
		timerText := formatDuration(elapsed)
		lines = append(lines, "")
		lines = append(lines, lipgloss.Place(width-4, 1, lipgloss.Center, lipgloss.Center, heroTimerStyle.Width(width-4).Render(timerText)))
		lines = append(lines, "")

		// Task info
		lines = append(lines, heroTaskStyle.Render("Task: "+entry.Text))
		lines = append(lines, heroTagStyle.Render("Start: "+startLocal.Format("2006-01-02 15:04")))

		// Tags
		tags := entry.Tags()
		if len(tags) == 0 {
			lines = append(lines, heroTagStyle.Render("Tags: (untagged)"))
		} else {
			tagColors := make([]string, len(tags))
			for i, tag := range tags {
				color := getTagColor(tag)
				tagColors[i] = lipgloss.NewStyle().Foreground(color).Render("#" + tag)
			}
			lines = append(lines, heroTagStyle.Render("Tags: "+strings.Join(tagColors, " ")))
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
