package components

import (
	"lazytime/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TagGroup represents entries grouped by tag (from aggregation).
type TagGroup struct {
	Tag       string
	Duration  time.Duration
	Entries   []storage.Entry
	Tasks     map[string]time.Duration
	TaskList  []TaskItem
}

// TaskItem represents a task with its duration.
type TaskItem struct {
	Text     string
	Duration time.Duration
	Start    time.Time
	End      time.Time
}

// RenderTree renders a hierarchical tree view of entries grouped by tag and task.
func RenderTree(groups []TagGroup, width, height int, treeTagStyle, treeTaskStyle, treeDurationStyle, boxStyle lipgloss.Style, getTagColor func(string) lipgloss.Color, formatDurationShort func(time.Duration) string) string {
	if len(groups) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No entries in this period."))
	}

	var lines []string
	maxLines := height - 2
	lineCount := 0

	for _, group := range groups {
		if lineCount >= maxLines {
			break
		}

		// Level 1: Tag with total duration
		tagColor := getTagColor(group.Tag)
		tagStyle := treeTagStyle.Copy().Foreground(tagColor)
		tagLine := "> " + tagStyle.Render(group.Tag)
		dots := strings.Repeat(".", max(0, width-len(tagLine)-len(formatDurationShort(group.Duration))-5))
		tagLine += " " + dots + " " + treeDurationStyle.Render(formatDurationShort(group.Duration))
		lines = append(lines, tagLine)
		lineCount++

		// Level 2: Tasks under this tag
		for _, task := range group.TaskList {
			if lineCount >= maxLines {
				break
			}

			// Format task with time range
			timeRange := task.Start.Format("15:04") + " - " + task.End.Format("15:04")

			taskLine := "  - " + treeTaskStyle.Render(task.Text)
			if len(taskLine)+len(timeRange)+len(formatDurationShort(task.Duration))+10 < width {
				taskLine += " (" + timeRange + ")"
			}
			dots := strings.Repeat(".", max(0, width-len(taskLine)-len(formatDurationShort(task.Duration))-5))
			taskLine += " " + dots + " " + treeDurationStyle.Render(formatDurationShort(task.Duration))

			lines = append(lines, taskLine)
			lineCount++
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Width(width).Height(height).Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
