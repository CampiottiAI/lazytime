package tui

import (
	"lazytime/storage"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// LaunchTUI initializes and launches the terminal UI using Bubbletea.
func LaunchTUI() error {
	m := NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// clampDuration calculates overlap duration within a time range.
// This is kept for backward compatibility with aggregation.go and components.
func clampDuration(entry storage.Entry, start, end, now time.Time) time.Duration {
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
