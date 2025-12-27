package components

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TagChartItem represents a tag with its duration for chart display.
type TagChartItem struct {
	Tag      string
	Duration time.Duration
	Percent  float64
}

// RenderTagChart renders a horizontal bar chart showing tag distribution.
func RenderTagChart(totals map[string]time.Duration, width, height int, chartBarStyle, chartLabelStyle, chartPercentStyle, boxStyle lipgloss.Style, getTagColor func(string) lipgloss.Color, formatDurationShort func(time.Duration) string) string {
	if len(totals) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No tags tracked."))
	}

	// Convert to slice and sort
	var items []TagChartItem
	var maxDuration time.Duration
	for tag, duration := range totals {
		items = append(items, TagChartItem{Tag: tag, Duration: duration})
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	// Calculate percentages
	for i := range items {
		if maxDuration > 0 {
			items[i].Percent = float64(items[i].Duration) / float64(maxDuration)
		}
	}

	// Sort by duration (descending)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Duration > items[j].Duration
	})

	// Limit to available height
	maxLines := height - 2
	if len(items) > maxLines {
		items = items[:maxLines]
	}

	var lines []string
	barWidth := width - 30 // Leave space for tag name and percentage

	for _, item := range items {
		filled := int(float64(barWidth) * item.Percent)
		if filled < 0 {
			filled = 0
		}
		if filled > barWidth {
			filled = barWidth
		}

		bar := ""
		for i := 0; i < filled; i++ {
			bar += "█"
		}

		tagColor := getTagColor(item.Tag)
		tagStyle := chartLabelStyle.Copy().Foreground(tagColor)
		tagName := tagStyle.Render(item.Tag)
		if len(tagName) > 15 {
			tagName = tagName[:12] + "..."
		}

		percentNum := int(item.Percent * 100)
		percentText := chartPercentStyle.Render(fmt.Sprintf("%d%%", percentNum))
		barStyled := chartBarStyle.Render(bar)

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Width(15).Render(tagName),
			barStyled,
			percentText,
		)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Width(width).Height(height).Render(content)
}
