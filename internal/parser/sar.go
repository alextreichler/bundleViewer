package parser

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type sarSection int

const (
	sectionUnknown sarSection = iota
	sectionCPU
	sectionMem
	sectionNetDev
	sectionNetTcp
	sectionNetETcp
)

// ParseSar parses a sar text file and extracts performance metrics across multiple sections.
func ParseSar(bundlePath string, logsOnly bool) (models.SarData, error) {
	candidates := []string{
		filepath.Join(bundlePath, "utils", "sar"),
		filepath.Join(bundlePath, "utils", "sar.txt"),
		filepath.Join(bundlePath, "sar.txt"),
		filepath.Join(bundlePath, "sysstat", "sar"),
	}

	var filePath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			filePath = c
			break
		}
	}

	if filePath == "" {
		return models.SarData{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return models.SarData{}, err
	}
	defer func() { _ = file.Close() }()

	// Use a map to aggregate metrics by timestamp string
	// Since sar files often have separate reports for CPU, Disk, Net, etc.
	entriesMap := make(map[string]*models.SarEntry)
	var timestamps []string

	scanner := bufio.NewScanner(file)
	currentSection := sectionUnknown
	baseDate := time.Now().Truncate(24 * time.Hour) // Fallback date

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Look for the main header to try and get the date
		// Linux 5.15.0 (hostname) 02/12/26 _x86_64_ (8 CPU)
		if fields[0] == "Linux" && len(fields) >= 4 {
			for _, f := range fields {
				if t, err := time.Parse("01/02/06", f); err == nil {
					baseDate = t
					break
				}
				if t, err := time.Parse("2006-01-02", f); err == nil {
					baseDate = t
					break
				}
			}
			continue
		}

		// Detect Section Headers
		newSection := detectSection(fields)
		if newSection != sectionUnknown {
			currentSection = newSection
			continue
		}

		// Parse Data Rows
		ts := fields[0]
		// Skip header rows that repeat timestamp
		if ts == "Average:" || ts == "Time" || ts == "HH:MM:SS" {
			continue
		}

		entry, exists := entriesMap[ts]
		if !exists {
			entry = &models.SarEntry{
				Timestamp: parseSarTimestamp(ts, baseDate),
			}
			entriesMap[ts] = entry
			timestamps = append(timestamps, ts)
		}

		switch currentSection {
		case sectionCPU:
			if fields[1] == "all" && len(fields) >= 8 {
				entry.CPUUser, _ = strconv.ParseFloat(fields[2], 64)
				entry.CPUSystem, _ = strconv.ParseFloat(fields[4], 64)
				entry.CPUIowait, _ = strconv.ParseFloat(fields[5], 64)
				entry.CPUIdle, _ = strconv.ParseFloat(fields[7], 64)
			}
		case sectionMem:
			// Time kbmemfree kbmemused  %memused kbbuffers  kbcached  kbcommit   %commit  kbactive   kbinact   kbdirty
			if len(fields) >= 4 {
				// Handle both newer and older sar formats for memory
				// Usually %memused is index 3 or 4
				for _, f := range fields {
					if strings.Contains(f, ".") {
						val, err := strconv.ParseFloat(f, 64)
						if err == nil && val <= 100.0 {
							entry.MemUsedPct = val
							break
						}
					}
				}
			}
		case sectionNetDev:
			// Time IFACE rxpck/s txpck/s rxkB/s txkB/s rxcmp/s txcmp/s rxmcst/s %ifutil
			if len(fields) >= 10 && fields[1] != "IFACE" {
				netEntry := models.SarNetDevEntry{
					IFace: fields[1],
				}
				netEntry.RxPck, _ = strconv.ParseFloat(fields[2], 64)
				netEntry.TxPck, _ = strconv.ParseFloat(fields[3], 64)
				netEntry.RxkB, _ = strconv.ParseFloat(fields[4], 64)
				netEntry.TxkB, _ = strconv.ParseFloat(fields[5], 64)
				netEntry.RxMcst, _ = strconv.ParseFloat(fields[8], 64)
				netEntry.IfUtil, _ = strconv.ParseFloat(fields[9], 64)
				entry.NetDev = append(entry.NetDev, netEntry)
			}
		case sectionNetTcp:
			// Time active/s passive/s iseg/s oseg/s
			if len(fields) >= 5 && fields[1] != "active/s" {
				entry.TcpActiveOpens, _ = strconv.ParseFloat(fields[1], 64)
				entry.TcpPassiveOpens, _ = strconv.ParseFloat(fields[2], 64)
				entry.TcpInSegs, _ = strconv.ParseFloat(fields[3], 64)
				entry.TcpOutSegs, _ = strconv.ParseFloat(fields[4], 64)
			}
		case sectionNetETcp:
			// Time atmptf/s estres/s retrans/s isegerr/s orsts/s
			if len(fields) >= 4 && fields[1] != "atmptf/s" {
				entry.TcpRetransRate, _ = strconv.ParseFloat(fields[3], 64)
				if len(fields) >= 5 {
					entry.TcpInErrs, _ = strconv.ParseFloat(fields[4], 64)
				}
			}
		}
	}

	// Convert map to sorted slice
	var result models.SarData
	// We sort the timestamps first to keep temporal order
	sort.Strings(timestamps)
	for _, ts := range timestamps {
		entry := entriesMap[ts]
		
		// Post-process derived metrics
		var peakTraffic float64
		for _, dev := range entry.NetDev {
			if dev.IFace == "lo" { continue }
			traffic := dev.RxkB + dev.TxkB
			if traffic > peakTraffic {
				peakTraffic = traffic
			}
		}
		entry.PeakThroughputMBps = peakTraffic / 1024.0

		result.Entries = append(result.Entries, *entry)
	}

	slog.Info("SAR data parsed", "entries", len(result.Entries))
	return result, nil
}

func detectSection(fields []string) sarSection {
	if len(fields) < 3 {
		return sectionUnknown
	}
	
	// CPU Header: Time CPU %user ...
	if fields[1] == "CPU" && fields[2] == "%user" {
		return sectionCPU
	}
	// Mem Header: Time kbmemfree kbmemused ...
	if fields[1] == "kbmemfree" || fields[1] == "kbmemused" {
		return sectionMem
	}
	// Net Dev Header: Time IFACE rxpck/s ...
	if fields[1] == "IFACE" && fields[2] == "rxpck/s" {
		return sectionNetDev
	}
	// Net TCP Header: Time active/s passive/s ...
	if fields[1] == "active/s" && fields[2] == "passive/s" {
		return sectionNetTcp
	}
	// Net ETCP Header: Time atmptf/s estres/s retrans/s ...
	if fields[1] == "atmptf/s" && fields[3] == "retrans/s" {
		return sectionNetETcp
	}

	return sectionUnknown
}

func parseSarTimestamp(ts string, baseDate time.Time) time.Time {
	// Handle 24h format
	if t, err := time.Parse("15:04:05", ts); err == nil {
		return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), t.Second(), 0, baseDate.Location())
	}
	// Handle 12h format (AM/PM) - fields would have been split by Fields()
	// but we only get the first part here. 
	// This is a limitation of the current simplified parser logic.
	// For now, return baseDate or try more common formats.
	return baseDate
}
