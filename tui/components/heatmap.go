package components

import (
	"lazytime/storage"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderWeekHeatmap renders a calendar heatmap for the week (7 days).
func RenderWeekHeatmap(entries []storage.Entry, now time.Time, width, height int, clampDuration func(storage.Entry, time.Time, time.Time, time.Time) time.Duration, boxStyle lipgloss.Style) string {
	tz := now.Location()
	today := now

	// Calculate week start (Monday)
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday-- // Monday = 0
	weekStart := today.AddDate(0, 0, -weekday)

	// Calculate daily totals
	dailyTotals := make([]time.Duration, 7)
	var maxDuration time.Duration

	for i := 0; i < 7; i++ {
		dayStart := weekStart.AddDate(0, 0, i)
		dayStartLocal := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, tz)
		dayEndLocal := dayStartLocal.AddDate(0, 0, 1)
		dayStartUTC := storage.ToUTC(dayStartLocal)
		dayEndUTC := storage.ToUTC(dayEndLocal)

		var total time.Duration
		for _, entry := range entries {
			total += clampDuration(entry, dayStartUTC, dayEndUTC, now)
		}
		dailyTotals[i] = total
		if total > maxDuration {
			maxDuration = total
		}
	}

	// Render squares
	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	var lines []string

	// Header
	header := "Week Heatmap"
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(header))
	lines = append(lines, "")

	// Squares row
	var squares []string
	for i, total := range dailyTotals {
		intensity := 0.0
		if maxDuration > 0 {
			intensity = float64(total) / float64(maxDuration)
		}

		// Choose color based on intensity
		var color lipgloss.Color
		if intensity == 0 {
			color = lipgloss.Color("#333333")
		} else if intensity < 0.25 {
			color = lipgloss.Color("#005500")
		} else if intensity < 0.5 {
			color = lipgloss.Color("#00aa00")
		} else if intensity < 0.75 {
			color = lipgloss.Color("#00ff00")
		} else {
			color = lipgloss.Color("#88ff88")
		}

		square := lipgloss.NewStyle().
			Background(color).
			Foreground(color).
			Width(2).
			Height(1).
			Render("██")

		squares = append(squares, square)
		if i < len(dayNames) {
			// Add day name below
			dayName := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(dayNames[i])
			squares = append(squares, "\n"+dayName)
		}
	}

	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, squares...))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Width(width).Height(height).Render(content)
}

// RenderMonthHeatmap renders a calendar heatmap for the month.
func RenderMonthHeatmap(entries []storage.Entry, now time.Time, width, height int, clampDuration func(storage.Entry, time.Time, time.Time, time.Time) time.Duration, boxStyle lipgloss.Style) string {
	// Simplified month view - show last 30 days
	tz := now.Location()
	today := now

	// Calculate daily totals for last 30 days
	dailyTotals := make([]time.Duration, 30)
	var maxDuration time.Duration

	for i := 0; i < 30; i++ {
		dayStart := today.AddDate(0, 0, -29+i)
		dayStartLocal := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, tz)
		dayEndLocal := dayStartLocal.AddDate(0, 0, 1)
		dayStartUTC := storage.ToUTC(dayStartLocal)
		dayEndUTC := storage.ToUTC(dayEndLocal)

		var total time.Duration
		for _, entry := range entries {
			total += clampDuration(entry, dayStartUTC, dayEndUTC, now)
		}
		dailyTotals[i] = total
		if total > maxDuration {
			maxDuration = total
		}
	}

	// Render grid (5 rows x 6 columns = 30 squares)
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Last 30 Days"))
	lines = append(lines, "")

	for row := 0; row < 5; row++ {
		var squares []string
		for col := 0; col < 6; col++ {
			idx := row*6 + col
			if idx >= 30 {
				break
			}

			total := dailyTotals[idx]
			intensity := 0.0
			if maxDuration > 0 {
				intensity = float64(total) / float64(maxDuration)
			}

			var color lipgloss.Color
			if intensity == 0 {
				color = lipgloss.Color("#333333")
			} else if intensity < 0.25 {
				color = lipgloss.Color("#005500")
			} else if intensity < 0.5 {
				color = lipgloss.Color("#00aa00")
			} else if intensity < 0.75 {
				color = lipgloss.Color("#00ff00")
			} else {
				color = lipgloss.Color("#88ff88")
			}

			square := lipgloss.NewStyle().
				Background(color).
				Foreground(color).
				Width(1).
				Height(1).
				Render("█")

			squares = append(squares, square)
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, squares...))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return boxStyle.Width(width).Height(height).Render(content)
}
