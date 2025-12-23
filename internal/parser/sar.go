package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// ParseSar parses a sar text file (often in utils/sar or just sar output files)
func ParseSar(bundlePath string, logsOnly bool) (models.SarData, error) {
	// Try to find a file that looks like sar output
	// Often named 'sar', 'sar.txt', or inside 'sysstat' dir
	candidates := []string{
		filepath.Join(bundlePath, "utils", "sar"),
		filepath.Join(bundlePath, "utils", "sar.txt"),
		filepath.Join(bundlePath, "sar.txt"),
	}

	var filePath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			filePath = c
			break
		}
	}

	if filePath == "" {
		return models.SarData{}, nil // Not found
	}

	file, err := os.Open(filePath)
	if err != nil {
		return models.SarData{}, err
	}
	defer func() { _ = file.Close() }()

	var data models.SarData
	scanner := bufio.NewScanner(file)
	
	// Very basic parser for "sar -u" output
	// Format: HH:MM:SS     CPU     %user     %nice   %system   %iowait    %steal     %idle
	// 12:00:01     all      0.10      0.00      0.20      0.00      0.00     99.70

	// We need to know the date to construct full timestamp, but sar output usually only has time.
	// We'll infer date from bundle timestamp or just use a dummy date for graphing (as time is relative).
	// Let's assume today's date or 0000-01-01.

	baseDate := time.Now().Truncate(24 * time.Hour) // Use current date for relative display

	headerFound := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		if !headerFound {
			if fields[1] == "CPU" && fields[2] == "%user" {
				headerFound = true
			}
			continue
		}

		// Parse Data Row
		// Time CPU %user %nice %system %iowait %steal %idle
		// 0    1   2     3     4       5       6      7
		if len(fields) >= 8 && (fields[1] == "all" || fields[1] == "CPU") { // "all" cores or header repetition
			// Parse Time
			t, err := time.Parse("15:04:05", fields[0])
			if err != nil {
				// Try AM/PM format
				t, err = time.Parse("03:04:05 PM", fields[0]+" "+fields[1]) // Logic gets complex with AM/PM fields splitting
				if err != nil {
					// Simplified: Just skip if fail
					continue
				}
			}
			
			// Combine with base date
			fullTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), t.Second(), 0, baseDate.Location())

			user, _ := strconv.ParseFloat(fields[2], 64)
			system, _ := strconv.ParseFloat(fields[4], 64)
			idle, _ := strconv.ParseFloat(fields[7], 64)

			data.Entries = append(data.Entries, models.SarEntry{
				Timestamp: fullTime,
				CPUUser:   user,
				CPUSystem: system,
				CPUIdle:   idle,
			})
		}
	}

	return data, nil
}
