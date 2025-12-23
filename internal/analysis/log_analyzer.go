package analysis

import (
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// GenerateFingerprint creates a generic signature for a log message
// Optimized single-pass implementation
func GenerateFingerprint(message string) string {
	var b strings.Builder
	b.Grow(len(message))

	n := len(message)
	for i := 0; i < n; i++ {
		c := message[i]

		// 1. Quoted Strings: "..." or '...'
		if c == '"' || c == '\'' {
			quote := c
			b.WriteString("<STR>")
			i++
			for i < n && message[i] != quote {
				i++
			}
			continue
		}

		// 2. Numbers and Hex/UUIDs
		// If we hit a digit, we check the "word"
		if isDigit(c) {
			// Look ahead to see if this is a hex string/UUID or just a number
			hasLetter := false
			
			// Scan until non-alphanumeric
			for i < n {
				curr := message[i]
				if isDigit(curr) {
					// ok
				} else if (curr >= 'a' && curr <= 'z') || (curr >= 'A' && curr <= 'Z') || curr == '-' {
					hasLetter = true
				} else if curr == '.' {
					// 123.456 (float) or 192.168.1.1 (IP)
				} else {
					// End of word (space, bracket, etc)
					break
				}
				i++
			}
			
			// Back up one since outer loop increments
			i--
			
			if hasLetter {
				// It was mixed alphanum (UUID, Hex, "2nd", IP?)
				b.WriteString("<ID>")
			} else {
				// Pure digits/dots
				b.WriteString("<NUM>")
			}
			continue
		}

		b.WriteByte(c)
	}
	return b.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// AnalyzeLogs groups logs by fingerprint
func AnalyzeLogs(logs []*models.LogEntry) []models.LogPattern {
	groups := make(map[string]*models.LogPattern)

	for _, entry := range logs {
		// Include Level and Component in signature to separate similar errors from different sources/levels
		// Or strictly use message? Usually grouping by message pattern is enough, but Level distinction is useful.
		// Let's fingerprint the message only, but store key as Level+Fingerprint

		fp := GenerateFingerprint(entry.Message)
		key := entry.Level + "|" + fp

		if _, exists := groups[key]; !exists {
			groups[key] = &models.LogPattern{
				Signature:   fp,
				Count:       0,
				Level:       entry.Level,
				SampleEntry: *entry,
				FirstSeen:   entry.Timestamp,
				LastSeen:    entry.Timestamp,
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
