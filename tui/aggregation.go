package tui

import (
	"sort"
	"strings"
	"time"

	"lazytime/storage"
)

// TagGroup represents entries grouped by tag.
type TagGroup struct {
	Tag       string
	Duration  time.Duration
	Entries   []storage.Entry
	Tasks     map[string]time.Duration // Task text -> duration
	TaskList  []TaskItem               // Sorted task items
}

// TaskItem represents a task with its duration.
type TaskItem struct {
	Text     string
	Duration time.Duration
	Start    time.Time
	End      time.Time
}

// GroupByTag groups entries by tag and calculates totals.
func GroupByTag(entries []storage.Entry, startUTC, endUTC, now time.Time) []TagGroup {
	tagMap := make(map[string]*TagGroup)

	for _, entry := range entries {
		duration := clampDuration(entry, startUTC, endUTC, now)
		if duration <= 0 {
			continue
		}

		tags := entry.Tags()
		if len(tags) == 0 {
			tags = []string{"(untagged)"}
		}

		for _, tag := range tags {
			group, exists := tagMap[tag]
			if !exists {
				group = &TagGroup{
					Tag:      tag,
					Duration: 0,
					Entries:  []storage.Entry{},
					Tasks:    make(map[string]time.Duration),
					TaskList: []TaskItem{},
				}
				tagMap[tag] = group
			}

			group.Duration += duration
			group.Entries = append(group.Entries, entry)

			// Group by task text (without tags)
			taskText := removeTags(entry.Text)
			group.Tasks[taskText] += duration
		}
	}

	// Convert to slice and sort by duration
	var groups []TagGroup
	for _, group := range tagMap {
		// Build sorted task list
		for taskText, taskDuration := range group.Tasks {
			// Find the first entry with this task text for start/end times
			var taskStart, taskEnd time.Time
			for _, entry := range group.Entries {
				if removeTags(entry.Text) == taskText {
					taskStart = entry.Start
					if entry.End != nil {
						taskEnd = *entry.End
					} else {
						taskEnd = now
					}
					break
				}
			}
			group.TaskList = append(group.TaskList, TaskItem{
				Text:     taskText,
				Duration: taskDuration,
				Start:    taskStart,
				End:      taskEnd,
			})
		}

		// Sort tasks by duration (descending)
		sort.Slice(group.TaskList, func(i, j int) bool {
			return group.TaskList[i].Duration > group.TaskList[j].Duration
		})

		groups = append(groups, *group)
	}

	// Sort groups by duration (descending)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Duration > groups[j].Duration
	})

	return groups
}

// GroupByTagAndTask creates a two-level hierarchy (tag -> task).
func GroupByTagAndTask(entries []storage.Entry, startUTC, endUTC, now time.Time) []TagGroup {
	return GroupByTag(entries, startUTC, endUTC, now)
}

// CalculateTagTotals calculates total duration per tag.
func CalculateTagTotals(entries []storage.Entry, startUTC, endUTC, now time.Time) map[string]time.Duration {
	totals := make(map[string]time.Duration)
	for _, entry := range entries {
		duration := clampDuration(entry, startUTC, endUTC, now)
		if duration <= 0 {
			continue
		}
		tags := entry.Tags()
		if len(tags) == 0 {
			tags = []string{"(untagged)"}
		}
		for _, tag := range tags {
			totals[tag] += duration
		}
	}
	return totals
}

// GetUniqueTags extracts all unique tags from entries.
func GetUniqueTags(entries []storage.Entry) []string {
	tagSet := make(map[string]bool)
	for _, entry := range entries {
		tags := entry.Tags()
		for _, tag := range tags {
			tagSet[strings.ToLower(tag)] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// FilterEntriesByRange filters entries that overlap with the given time range.
func FilterEntriesByRange(entries []storage.Entry, startUTC, endUTC, now time.Time) []storage.Entry {
	var filtered []storage.Entry
	for _, entry := range entries {
		if clampDuration(entry, startUTC, endUTC, now) > 0 {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

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

