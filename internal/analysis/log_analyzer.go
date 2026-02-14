package analysis

import (
	"sort"

	"github.com/alextreichler/bundleViewer/internal/logutil"
	"github.com/alextreichler/bundleViewer/internal/models"
)

// AnalyzeLogs groups logs by fingerprint
func AnalyzeLogs(logs []*models.LogEntry) []models.LogPattern {
	groups := make(map[string]*models.LogPattern)

	for _, entry := range logs {
		fp := logutil.GenerateFingerprint(entry.Message)
		key := entry.Level + "|" + fp

		if _, exists := groups[key]; !exists {
			groups[key] = &models.LogPattern{
				Signature:   fp,
				Count:       0,
				Level:       entry.Level,
				SampleEntry: *entry,
				FirstSeen:   entry.Timestamp,
				LastSeen:    entry.Timestamp,
				Insight:     logutil.GetInsight(entry.Message),
			}
		}

		g := groups[key]
		g.Count++
		if entry.Timestamp.Before(g.FirstSeen) {
			g.FirstSeen = entry.Timestamp
		}
		if entry.Timestamp.After(g.LastSeen) {
			g.LastSeen = entry.Timestamp
		}
	}

	// Convert map to slice
	var patterns []models.LogPattern
	for _, p := range groups {
		patterns = append(patterns, *p)
	}

	// Sort by Count (Desc)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}
