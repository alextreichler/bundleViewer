package parser

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

// parseSizeToBytes converts a human-readable size string (e.g., "1.2M", "3G", "500K") to bytes.
func parseSizeToBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "0" {
		return 0, nil
	}

	var multiplier int64 = 1

	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	} else if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "T")
	} else if strings.HasSuffix(sizeStr, "B") { // Handle explicit 'B' for bytes
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	parsedValue, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size value '%s': %w", sizeStr, err)
	}

	return int64(parsedValue * float64(multiplier)), nil
}

// tryParseStandardLog attempts to parse a log line using fast string operations.
// Format: LEVEL  YYYY-MM-DD HH:MM:SS,mmm [shard X] component - message
func tryParseStandardLog(line string) (level, ts, shard, comp, msg string, ok bool) {
	if len(line) < 30 {
		return "", "", "", "", "", false
	}

	// 1. Level (INFO, WARN, ERROR, DEBUG, TRACE)
	idx := strings.IndexByte(line, ' ')
	if idx < 4 || idx > 5 {
		return "", "", "", "", "", false
	}
	level = line[:idx]
	switch level {
	case "INFO", "WARN", "ERROR", "DEBUG", "TRACE":
		// Valid level
	default:
		return "", "", "", "", "", false
	}

	// 2. Timestamp
	rest := line[idx:]
	// Skip spaces
	tsIdx := 0
	for tsIdx < len(rest) && rest[tsIdx] == ' ' {
		tsIdx++
	}
	rest = rest[tsIdx:]
	if len(rest) < 23 || rest[4] != '-' || rest[7] != '-' {
		return "", "", "", "", "", false
	}
	ts = rest[:23]
	rest = rest[23:]

	// 3. Shard (optional)
	rest = strings.TrimLeft(rest, " ")
	if strings.HasPrefix(rest, "[shard") {
		endIdx := strings.IndexByte(rest, ']')
		if endIdx != -1 {
			shard = strings.TrimSpace(rest[6:endIdx])
			rest = rest[endIdx+1:]
		}
	}

	// 4. Component and Message
	rest = strings.TrimLeft(rest, " ")
	hyphenIdx := strings.Index(rest, " - ")
	if hyphenIdx != -1 {
		comp = strings.TrimSpace(rest[:hyphenIdx])
		msg = strings.TrimSpace(rest[hyphenIdx+3:])
	} else {
		// Fallback to first space
		spaceIdx := strings.IndexByte(rest, ' ')
		if spaceIdx != -1 {
			comp = strings.TrimSpace(rest[:spaceIdx])
			msg = strings.TrimSpace(rest[spaceIdx+1:])
		} else {
			comp = strings.TrimSpace(rest)
		}
	}

	return level, ts, shard, comp, msg, true
}

// tryParseK8sLog attempts to parse a Kubernetes log line using fast string operations.
// Format: YYYY-MM-DDTHH:MM:SSZ ... level=X ...
func tryParseK8sLog(line string) (ts, level, comp, msg string, ok bool) {
	if len(line) < 20 {
		return "", "", "", "", false
	}

	// 1. Timestamp (ISO8601 at start)
	// Expect 2006-01-02T15:04:05Z (20 chars) or similar
	if line[10] != 'T' || line[19] != 'Z' {
		return "", "", "", "", false
	}
	ts = line[:20]
	rest := line[20:]

	// 2. Level parsing (look for level=...)
	// Scan for "level="
	level = "INFO" // Default
	levelIdx := strings.Index(rest, "level=")
	if levelIdx != -1 {
		// Extract value
		start := levelIdx + 6
		end := start
		// Find end of value (space or quote)
		if start < len(rest) && rest[start] == '"' {
			start++
			end = start
			for end < len(rest) && rest[end] != '"' {
				end++
			}
		} else {
			for end < len(rest) && rest[end] != ' ' {
				end++
			}
		}
		if start < len(rest) {
			level = strings.ToUpper(rest[start:end])
		}
	}

	// 3. Component & Message
	comp = "k8s-sidecar"
	msg = strings.TrimSpace(rest)

	return ts, level, comp, msg, true
}

// tryParseCSVLog attempts to parse a generic CSV log line.
// It heuristically looks for a timestamp and a level.
func tryParseCSVLog(line string) (ts, level, comp, msg string, ok bool) {
	r := csv.NewReader(strings.NewReader(line))
	r.FieldsPerRecord = -1 // Variable fields
	record, err := r.Read()
	if err != nil || len(record) < 2 {
		return "", "", "", "", false
	}

	// Heuristic 1: Look for Timestamp
	// Common formats: ISO8601, RFC3339
	// We'll iterate fields to find something that looks like a date
	tsIdx := -1
	for i, field := range record {
		if len(field) >= 19 && (field[10] == 'T' || field[10] == ' ') {
			// Weak check for 2006-01-02...
			if field[4] == '-' && field[7] == '-' {
				tsIdx = i
				ts = field
				break
			}
		}
	}

	if tsIdx == -1 {
		return "", "", "", "", false
	}

	// Heuristic 2: Look for Level (INFO, WARN, etc.)
	// Usually short, uppercase
	level = "INFO"
	levelIdx := -1
	for i, field := range record {
		if i == tsIdx {
			continue
		}
		upper := strings.ToUpper(field)
		switch upper {
		case "INFO", "WARN", "WARNING", "ERROR", "ERR", "DEBUG", "TRACE", "FATAL", "CRITICAL":
			level = upper
			levelIdx = i
		}
		if levelIdx != -1 {
			break
		}
	}

	// Heuristic 3: Component & Message
	// Message is likely the longest remaining field
	comp = "csv-export"
	longestLen := 0
	
	for i, field := range record {
		if i == tsIdx || i == levelIdx {
			continue
		}
		if len(field) > longestLen {
			msg = field
			longestLen = len(field)
		}
	}

	if msg == "" {
		// If no message found (weird), join the rest
		var parts []string
		for i, field := range record {
			if i != tsIdx && i != levelIdx {
				parts = append(parts, field)
			}
		}
		msg = strings.Join(parts, " ")
	}

	return ts, level, comp, msg, true
}

// ParseLogs reads and parses all redpanda log files and stores them in the provided store.
func ParseLogs(bundlePath string, s store.Store, logsOnly bool, p *models.ProgressTracker) error {
	var allLogFiles []string

	if logsOnly {
		// In Logs Only mode, assume any file in the root is potentially a log file
		entries, err := os.ReadDir(bundlePath)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					// Skip likely non-log system files if obvious, but user requested "any file"
					// We'll trust the user but maybe skip dotfiles like .DS_Store
					if strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					allLogFiles = append(allLogFiles, filepath.Join(bundlePath, entry.Name()))
				}
			}
		}
	} else {
		// In Full Bundle mode, be strict to avoid parsing config/status text files as logs
		mainLogFile := filepath.Join(bundlePath, "redpanda.log")
		if _, err := os.Stat(mainLogFile); err == nil {
			allLogFiles = append(allLogFiles, mainLogFile)
		}
	}

	// 2. Recursively find *.log and *.txt files in logs/
	dirsToWalk := []string{
		filepath.Join(bundlePath, "logs"),
	}

	for _, dir := range dirsToWalk {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				slog.Error("Error walking directory", "path", path, "error", err)
				return nil // Don't stop the walk on error for one file
			}
			if d.IsDir() {
				return nil // Skip directories
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".log" || ext == ".txt" || ext == ".csv" {
				allLogFiles = append(allLogFiles, path)
			}
			return nil
		})
		if err != nil {
			slog.Error("Error during filepath.WalkDir", "path", dir, "error", err)
		}
	}

	// Use a map to store unique file paths and avoid processing the same file multiple times
	uniqueFiles := make(map[string]struct{})
	var distinctFiles []string
	for _, file := range allLogFiles {
		if _, exists := uniqueFiles[file]; !exists {
			uniqueFiles[file] = struct{}{}
			distinctFiles = append(distinctFiles, file)
		}
	}

	// Channel for sending batches to the DB writer
	// Buffer size allow parsers to proceed while DB is writing
	logBatchChan := make(chan []*models.LogEntry, 20)
	
	// Error channel to collect errors from goroutines
	errChan := make(chan error, len(distinctFiles)+1)

	var parserWg sync.WaitGroup
	var writerWg sync.WaitGroup

	slog.Info("Starting log parsing", "file_count", len(distinctFiles))

	// Heartbeat for UI liveliness
	heartbeatDone := make(chan struct{})
	if p != nil {
		go func() {
			ticker := time.NewTicker(20 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-heartbeatDone:
					return
				case <-ticker.C:
					p.SetStatus("Still processing logs... (not stuck)")
				}
			}
		}()
	}

	// Start DB Writer Goroutine
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for batch := range logBatchChan {
			if err := s.BulkInsertLogs(batch); err != nil {
				slog.Error("Error bulk inserting logs", "error", err)
				// Don't block if errChan is full
				select {
				case errChan <- fmt.Errorf("db write error: %w", err):
				default:
					slog.Warn("Error channel full, dropping error", "err", err)
				}
				return // Stop writing on error to avoid deadlock
			}
		}
	}()

	// Removed slow Regex compilations
	layout := "2006-01-02 15:04:05,000"

	const batchSize = 10000 // Increased batch size for better throughput

	// Limit concurrent file parsing to number of CPUs to avoid OOM
	sem := make(chan struct{}, runtime.NumCPU())

	for _, file := range distinctFiles {
		parserWg.Add(1)
		// Acquire semaphore
		sem <- struct{}{}
		go func(filePath string) {
			defer parserWg.Done()
			defer func() { <-sem }() // Release semaphore
			if p != nil {
				p.SetStatus(fmt.Sprintf("Parsing logs: %s", filepath.Base(filePath)))
			}
			slog.Info("Parsing file", "path", filePath)

			nodeName := ""
			baseName := filepath.Base(filePath)
			if baseName == "redpanda.log" {
				nodeName = "redpanda"
			} else {
				nodeName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
			}

			f, err := os.Open(filePath)
			if err != nil {
				slog.Error("Error opening log file", "path", filePath, "error", err)
				errChan <- fmt.Errorf("failed to open %s: %w", filePath, err)
				return
			}
			defer func() { _ = f.Close() }()

			// Pre-allocate batch slice to reduce re-allocations
			batch := make([]*models.LogEntry, 0, batchSize)
			
			scanner := bufio.NewScanner(f)
			const maxCapacity = 5 * 1024 * 1024 // 5MB
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			var currentEntry *models.LogEntry
			lineNumber := 0

			for scanner.Scan() {
				lineNumber++
				if p != nil && lineNumber%20000 == 0 {
					p.SetStatus(fmt.Sprintf("Parsing logs: %s (%d lines processed)", filepath.Base(filePath), lineNumber))
				}
				line := scanner.Text()

				if len(line) > 50000 {
					continue // Skip massive lines
				}

				// Try matching with regex - REPLACED with Fast String Parsing

				isCSV := strings.HasSuffix(filePath, ".csv")

				// 0. CSV PATH
				if isCSV {
					cTs, cLevel, cComp, cMsg, cOk := tryParseCSVLog(line)
					if cOk {
						if currentEntry != nil {
							batch = append(batch, currentEntry)
							if len(batch) >= batchSize {
								logBatchChan <- batch
								batch = make([]*models.LogEntry, 0, batchSize)
							}
						}

						// Try parsing timestamp formats
						timestamp, err := time.Parse(time.RFC3339, cTs)
						if err != nil {
							timestamp, err = time.Parse(time.RFC3339Nano, cTs)
						}
						if err != nil {
							timestamp, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", cTs) // Google Cloud format sometimes
						}

						currentEntry = &models.LogEntry{
							Timestamp:  timestamp,
							Level:      cLevel,
							Node:       nodeName,
							Component:  cComp,
							Message:    cMsg,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
						continue
					}
				}

				// 1. FAST PATH: Try parsing standard Redpanda log format without regex
				if !isCSV {
					fLevel, fTs, fShard, fComp, fMsg, ok := tryParseStandardLog(line)
					if ok {
						if currentEntry != nil {
							batch = append(batch, currentEntry)
							if len(batch) >= batchSize {
								logBatchChan <- batch
								batch = make([]*models.LogEntry, 0, batchSize)
							}
						}

						// Normalize timestamp
						fTs = strings.Replace(fTs, ".", ",", 1)
						timestamp, _ := time.Parse(layout, fTs)

						// Clean up shard info
						if idx := strings.Index(fShard, ":"); idx != -1 {
							fShard = fShard[:idx]
						}

						currentEntry = &models.LogEntry{
							Timestamp:  timestamp,
							Level:      fLevel,
							Node:       nodeName,
							Shard:      fShard,
							Component:  fComp,
							Message:    fMsg,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
						continue
					}
				}

				// 2. K8S PATH: Try parsing K8s/JSON logs without regex
				if !isCSV {
					kTs, kLevel, kComp, kMsg, kOk := tryParseK8sLog(line)
					if kOk {
						if currentEntry != nil {
							batch = append(batch, currentEntry)
							if len(batch) >= batchSize {
								logBatchChan <- batch
								batch = make([]*models.LogEntry, 0, batchSize)
							}
						}
						
						timestamp, _ := time.Parse(time.RFC3339, kTs)

						currentEntry = &models.LogEntry{
							Timestamp:  timestamp,
							Level:      kLevel,
							Node:       nodeName,
							Component:  kComp,
							Message:    kMsg,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
						continue
					}
				}

				// 3. JSON Fallback
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
					var jsonLog map[string]interface{}
					if json.Unmarshal([]byte(trimmed), &jsonLog) == nil {
						lvl, _ := jsonLog["level"].(string)
						if lvl == "" { lvl = "INFO" }
						lvl = strings.ToUpper(lvl)

						tsStr, _ := jsonLog["ts"].(string)
						if tsStr == "" { tsStr, _ = jsonLog["timestamp"].(string) }
						ts, err := time.Parse(time.RFC3339, tsStr)
						if err != nil {
							ts, _ = time.Parse("2006-01-02T15:04:05.999Z07:00", tsStr)
						}

						msg, _ := jsonLog["msg"].(string)
						if msg == "" { msg, _ = jsonLog["message"].(string) }
						if msg == "" { msg = trimmed }

						comp, _ := jsonLog["logger"].(string)
						if comp == "" { comp, _ = jsonLog["component"].(string) }
						if comp == "" { comp = "sidecar" }

						if currentEntry != nil {
							batch = append(batch, currentEntry)
							if len(batch) >= batchSize {
								logBatchChan <- batch
								batch = make([]*models.LogEntry, 0, batchSize)
							}
						}
						
						currentEntry = &models.LogEntry{
							Timestamp:  ts,
							Level:      lvl,
							Node:       nodeName,
							Component:  comp,
							Message:    msg,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
						continue
					}
				}

				// 4. Multiline Handling
				if currentEntry != nil {
					currentEntry.Message += "\n" + line
					currentEntry.Raw += "\n" + line
				} else {
					// Fallback for completely unknown lines (assume STDOUT INFO)
					currentEntry = &models.LogEntry{
						Timestamp:  time.Time{},
						Level:      "INFO",
						Node:       nodeName,
						Component:  "stdout",
						Message:    line,
						Raw:        line,
						LineNumber: lineNumber,
						FilePath:   filePath,
					}
				}
			}

			if currentEntry != nil {
				batch = append(batch, currentEntry)
			}

			if len(batch) > 0 {
				logBatchChan <- batch
			}

			if err := scanner.Err(); err != nil {
				slog.Error("Error scanning log file", "path", filePath, "error", err)
				errChan <- fmt.Errorf("error reading %s: %w", filePath, err)
			}
		}(file)
	}

	// Wait for all parsers to finish
	parserWg.Wait()
	// Then close the channel to signal DB writer to stop
	close(logBatchChan)
	// Wait for DB writer to finish
	writerWg.Wait()
	
	close(errChan)
	
	// Stop heartbeat
	if p != nil {
		close(heartbeatDone)
	}

	// Return first error if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseDuOutput reads and parses the du.txt file.
func ParseDuOutput(bundlePath string, logsOnly bool) ([]models.DuEntry, error) {
	filePath := filepath.Join(bundlePath, "utils", "du.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []models.DuEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "du:") { // Skip empty lines and du error messages
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			if !logsOnly {
				slog.Warn("Skipping malformed du.txt line", "line", line)
			}
			continue
		}

		size, err := parseSizeToBytes(parts[0])
		if err != nil {
			if !logsOnly {
				slog.Warn("Failed to parse size for du.txt entry", "line", line, "error", err)
			}
			continue
		}

		path := parts[1]
		entries = append(entries, models.DuEntry{Size: size, Path: path})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading du.txt: %w", err)
	}

	return entries, nil
}

// KafkaFullData encapsulates all data parsed from kafka.json.
type KafkaFullData struct {
	Metadata       models.KafkaMetadataResponse
	TopicConfigs   models.TopicConfigsResponse
	BrokerConfigs  models.BrokerConfigsResponse
	ConsumerGroups map[string]models.ConsumerGroup
}

// ParseKafkaJSON reads and parses the entire kafka.json file.
func ParseKafkaJSON(bundlePath string) (KafkaFullData, error) {
	filePath := filepath.Join(bundlePath, "kafka.json")
	f, err := os.Open(filePath)
	if err != nil {
		return KafkaFullData{}, err
	}
	defer func() { _ = f.Close() }()

	fullData := KafkaFullData{
		ConsumerGroups: make(map[string]models.ConsumerGroup),
	}

	// Map to temporarily hold group commits before joining with groups
	groupCommits := make(map[string]map[string]map[int]*models.GroupPartitionOffset)
	highWatermarks := make(map[string]map[int]int64)

	dec := json.NewDecoder(f)

	// Check for opening bracket
	if token, err := dec.Token(); err != nil || token != json.Delim('[') {
		return KafkaFullData{}, fmt.Errorf("expected array start in kafka.json")
	}

	for dec.More() {
		var item struct {
			Name     string          `json:"Name"`
			Response json.RawMessage `json:"Response"`
		}
		if err := dec.Decode(&item); err != nil {
			return KafkaFullData{}, fmt.Errorf("error decoding item in kafka.json: %w", err)
		}

		switch {
		case item.Name == "metadata":
			var metadata models.KafkaMetadataResponse
			if err := json.Unmarshal(item.Response, &metadata); err == nil {
				fullData.Metadata = metadata
				slog.Debug("Parsed metadata section from kafka.json")
			}
		case item.Name == "topic_configs":
			var topicConfigs models.TopicConfigsResponse
			if err := json.Unmarshal(item.Response, &topicConfigs); err == nil {
				fullData.TopicConfigs = topicConfigs
				slog.Debug("Parsed topic_configs section from kafka.json")
			}
		case item.Name == "broker_configs":
			var brokerConfigs models.BrokerConfigsResponse
			if err := json.Unmarshal(item.Response, &brokerConfigs); err == nil {
				fullData.BrokerConfigs = brokerConfigs
				slog.Debug("Parsed broker_configs section from kafka.json")
			}
		case item.Name == "groups":
			var groups map[string]models.ConsumerGroup
			if err := json.Unmarshal(item.Response, &groups); err == nil {
				for k, v := range groups {
					fullData.ConsumerGroups[k] = v
				}
				slog.Debug("Parsed groups section from kafka.json", "count", len(groups))
			} else {
				slog.Error("Failed to unmarshal groups section", "error", err)
			}
		case item.Name == "high_watermarks":
			var hwm map[string]map[int]models.HighWatermark
			if err := json.Unmarshal(item.Response, &hwm); err == nil {
				for topic, partitions := range hwm {
					if highWatermarks[topic] == nil {
						highWatermarks[topic] = make(map[int]int64)
					}
					for partID, hw := range partitions {
						highWatermarks[topic][partID] = hw.Offset
					}
				}
				slog.Debug("Parsed high_watermarks section from kafka.json")
			}
		case strings.HasPrefix(item.Name, "group_commits_"):
			groupID := strings.TrimPrefix(item.Name, "group_commits_")
			var commits map[string]map[int]*models.GroupPartitionOffset
			if err := json.Unmarshal(item.Response, &commits); err == nil {
				groupCommits[groupID] = commits
				slog.Debug("Parsed group_commits section from kafka.json", "groupID", groupID)
			}
		}
	}

	// Check for closing bracket
	if token, err := dec.Token(); err != nil || token != json.Delim(']') {
		return KafkaFullData{}, fmt.Errorf("expected array end in kafka.json")
	}

	// Join commits with groups and calculate lag
	for groupID, commits := range groupCommits {
		group, ok := fullData.ConsumerGroups[groupID]
		if !ok {
			// Some groups might have commits but not be in the "groups" list (e.g. inactive)
			group = models.ConsumerGroup{Group: groupID, State: "Unknown"}
		}

		var totalGroupLag int64
		for topic, partitions := range commits {
			var topicLag int64
			for _, part := range partitions {
				if hwmTopic, ok := highWatermarks[topic]; ok {
					if hwmOffset, ok := hwmTopic[part.Partition]; ok {
						part.Lag = hwmOffset - part.At
						if part.Lag < 0 {
							part.Lag = 0 // Offset could be ahead of HWM in some race conditions or if HWM is stale
						}
						topicLag += part.Lag
					}
				}
			}
			group.Offsets = append(group.Offsets, models.GroupTopicOffset{
				Topic:      topic,
				Partitions: partitions,
				TopicLag:   topicLag,
			})
			totalGroupLag += topicLag
		}
		group.TotalLag = totalGroupLag
		fullData.ConsumerGroups[groupID] = group
	}

	return fullData, nil
}

// ParseKafkaMetadata reads and parses the kafka.json file.
func ParseKafkaMetadata(bundlePath string) (models.KafkaMetadataResponse, error) {
	fd, err := ParseKafkaJSON(bundlePath)
	if err != nil {
		return models.KafkaMetadataResponse{}, err
	}
	return fd.Metadata, nil
}

// ParseTopicConfigs reads and parses the topic_configs section from kafka.json.
func ParseTopicConfigs(bundlePath string) (models.TopicConfigsResponse, error) {
	fd, err := ParseKafkaJSON(bundlePath)
	if err != nil {
		return nil, err
	}
	return fd.TopicConfigs, nil
}


// ParseHealthOverview reads and parses the health_overview.json file.
func ParseHealthOverview(bundlePath string) (models.HealthOverview, error) {
	filePath := filepath.Join(bundlePath, "admin", "health_overview.json") // Assuming it's in admin folder
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.HealthOverview{}, err
	}

	var healthOverview models.HealthOverview
	err = json.Unmarshal(data, &healthOverview)
	if err != nil {
		slog.Error("Error unmarshaling health_overview.json", "error", err)
		return models.HealthOverview{}, err
	}
	return healthOverview, nil
}

// ParsePartitionBalancerStatus reads and parses the partition_balancer_status.json file.
func ParsePartitionBalancerStatus(bundlePath string) (models.PartitionBalancerStatus, error) {
	filePath := filepath.Join(bundlePath, "admin", "partition_balancer_status.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.PartitionBalancerStatus{}, err
	}

	var status models.PartitionBalancerStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		slog.Error("Error unmarshaling partition_balancer_status.json", "error", err)
		return models.PartitionBalancerStatus{}, err
	}
	return status, nil
}

// ParsePartitionLeaders reads and parses a partition_leader_table_...json file.
func ParsePartitionLeaders(bundlePath string, logsOnly bool) ([]models.PartitionLeader, error) {
	pattern := filepath.Join(bundlePath, "admin", "partition_leader_table_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		if !logsOnly {
			slog.Warn("No partition_leader_table file found.")
		}
		return []models.PartitionLeader{}, nil
	}

	// Try to parse each file until one succeeds
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			if !logsOnly {
				slog.Warn("Failed to read", "file", file, "error", err)
			}
			continue
		}

		var leaders []models.PartitionLeader
		if err := json.Unmarshal(data, &leaders); err == nil && len(leaders) > 0 {
			return leaders, nil
		} else {
			if !logsOnly {
				slog.Warn("Failed to parse", "file", file, "error", err)
			}
		}
	}

	return nil, fmt.Errorf("failed to parse any partition_leader_table file")
}

// ParseClusterPartitions reads and parses the cluster_partitions.json file.
func ParseClusterPartitions(bundlePath string) ([]models.ClusterPartition, error) {
	filePath := filepath.Join(bundlePath, "admin", "cluster_partitions.json")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var partitions []models.ClusterPartition
	
	dec := json.NewDecoder(f)
	
	// Check for opening bracket
	if token, err := dec.Token(); err != nil || token != json.Delim('[') {
		return nil, fmt.Errorf("expected array start in cluster_partitions.json")
	}

	for dec.More() {
		var p models.ClusterPartition
		if err := dec.Decode(&p); err != nil {
			return nil, err
		}
		partitions = append(partitions, p)
	}

	// Check for closing bracket
	if token, err := dec.Token(); err != nil || token != json.Delim(']') {
		return nil, fmt.Errorf("expected array end in cluster_partitions.json")
	}

	return partitions, nil
}

// ParsedFile holds the filename and the parsed data of a file.
type ParsedFile struct {
	FileName string
	Data     interface{}
	Error    error
}

// ParseSpecificFiles reads and parses a list of specific files.
func ParseSpecificFiles(baseDir string, filenames []string) ([]ParsedFile, error) {
	var parsedFiles []ParsedFile

	for _, filename := range filenames {
		filePath := filepath.Join(baseDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			parsedFiles = append(parsedFiles, ParsedFile{
				FileName: filename,
				Error:    fmt.Errorf("file not found"),
			})
			continue
		}

		var parsedData interface{}
		if strings.HasSuffix(filename, ".json") {
			err = json.Unmarshal(data, &parsedData)
			if err != nil {
				// If JSON unmarshaling fails, treat it as plain text but log an error
				parsedFiles = append(parsedFiles, ParsedFile{
					FileName: filename,
					Data:     string(data),
					Error:    fmt.Errorf("failed to parse JSON: %w", err),
				})
				continue
			}
		} else {
			// Treat as plain text for other extensions (including .txt, .yaml, etc.)
			parsedData = string(data)
		}

		parsedFiles = append(parsedFiles, ParsedFile{
			FileName: filename,
			Data:     parsedData,
		})
	}

	return parsedFiles, nil
}

// ParseAllFiles reads and parses all files with specified extensions in a directory.
func ParseAllFiles(dir string, fileExtensions []string) ([]ParsedFile, error) {
	var parsedFiles []ParsedFile

	for _, ext := range fileExtensions {
		pattern := filepath.Join(dir, "*"+ext)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				parsedFiles = append(parsedFiles, ParsedFile{
					FileName: filepath.Base(file),
					Error:    fmt.Errorf("file not found"),
				})
				continue
			}

			var parsedData interface{}
			if strings.HasSuffix(file, ".json") {
				err = json.Unmarshal(data, &parsedData)
				if err != nil {
					// If JSON unmarshaling fails, treat it as plain text but log an error
					parsedFiles = append(parsedFiles, ParsedFile{
						FileName: filepath.Base(file),
						Data:     string(data),
						Error:    fmt.Errorf("failed to parse JSON: %w", err),
					})
					continue
				}
			} else {
				// Treat as plain text for other extensions
				parsedData = string(data)
			}

			parsedFiles = append(parsedFiles, ParsedFile{
				FileName: filepath.Base(file),
				Data:     parsedData,
			})
		}
	}

	return parsedFiles, nil
}

// ParseUnameInfo reads and parses the uname.txt file to extract system information.
func ParseUnameInfo(bundlePath string) (models.UnameInfo, error) {
	filePath := filepath.Join(bundlePath, "utils", "uname.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.UnameInfo{}, err
	}

	// uname -a output is typically:
	// Linux hostname 5.15.0-100-generic #110-Ubuntu SMP ... x86_64 ...
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		info := models.UnameInfo{
			KernelName:    fields[0],
			Hostname:      fields[1],
			KernelRelease: fields[2],
		}
		
		// Set a default OS if it's Linux
		if info.KernelName == "Linux" {
			info.OperatingSystem = "Linux"
		}

		// Infer Distribution from KernelRelease pattern
		release := info.KernelRelease
		if strings.Contains(release, ".el9") {
			info.Distro = "RHEL/CentOS 9"
		} else if strings.Contains(release, ".el8") {
			info.Distro = "RHEL/CentOS 8"
		} else if strings.Contains(release, ".el7") {
			info.Distro = "RHEL/CentOS 7"
		} else if strings.Contains(release, "Ubuntu") || strings.Contains(info.KernelVersion, "Ubuntu") {
			info.Distro = "Ubuntu"
		} else if strings.Contains(release, "Debian") || strings.Contains(info.KernelVersion, "Debian") {
			info.Distro = "Debian"
		} else if strings.Contains(release, "amzn") {
			if strings.Contains(release, "amzn2") {
				info.Distro = "Amazon Linux 2"
			} else if strings.Contains(release, "amzn2023") {
				info.Distro = "Amazon Linux 2023"
			} else {
				info.Distro = "Amazon Linux"
			}
		} else if strings.Contains(release, "fc") {
			info.Distro = "Fedora"
		} else if strings.Contains(release, "arch") {
			info.Distro = "Arch Linux"
		} else if strings.Contains(release, "gentoo") {
			info.Distro = "Gentoo"
		} else if strings.Contains(release, "alpine") {
			info.Distro = "Alpine Linux"
		} else if strings.Contains(release, "rocky") {
			info.Distro = "Rocky Linux"
		} else if strings.Contains(release, "alma") {
			info.Distro = "AlmaLinux"
		}
		
		// Detect if running in a container/K8s
		// Heuristics:
		// 1. Hostname matches standard K8s pod pattern (e.g. name-random-hash)
		// 2. Hostname matches statefulset pattern (e.g. name-0)
		hostname := info.Hostname
		podPattern := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*-[0-9]+$`) // statefulset
		if podPattern.MatchString(hostname) || strings.HasPrefix(hostname, "redpanda-") {
			info.IsContainer = true
		}
		
		// Detect Cloud Provider
		if strings.Contains(hostname, ".aws") || strings.Contains(hostname, "compute.internal") {
			info.CloudProvider = "AWS"
		} else if strings.Contains(hostname, ".google") || strings.Contains(hostname, "gce") {
			info.CloudProvider = "GCP"
		} else if strings.Contains(hostname, ".azure") {
			info.CloudProvider = "Azure"
		}

		// Try to find architecture (usually last or second to last, e.g. x86_64)
		// And OS (GNU/Linux)
		for i, field := range fields {
			if field == "x86_64" || field == "aarch64" || field == "arm64" || field == "x86" || field == "i386" || field == "i686" {
				info.Machine = field
			}
			if field == "GNU/Linux" {
				info.OperatingSystem = field
			}
			// Version often starts with # and contains build info, spreading across multiple fields
			if strings.HasPrefix(field, "#") && i > 2 {
				// Reconstruct version string roughly until we hit arch or end
				// This is heuristic as uname output varies by OS
				end := len(fields)
				for j := i; j < len(fields); j++ {
					if fields[j] == "x86_64" || fields[j] == "aarch64" || fields[j] == "arm64" || fields[j] == "GNU/Linux" {
						end = j
						break
					}
				}
				info.KernelVersion = strings.Join(fields[i:end], " ")
				
				// Check KernelVersion for distro hints if release didn't have any
				if info.Distro == "" {
					if strings.Contains(info.KernelVersion, "Ubuntu") {
						info.Distro = "Ubuntu"
					} else if strings.Contains(info.KernelVersion, "Debian") {
						info.Distro = "Debian"
					}
				}
			}
		}
		
		return info, nil
	}

	return models.UnameInfo{}, fmt.Errorf("could not parse uname.txt: insufficient fields")
}

// ParseClusterConfig reads and parses the cluster_config.json file, generating documentation links.
func ParseClusterConfig(bundlePath string) ([]models.ClusterConfigEntry, error) {
	filePath := filepath.Join(bundlePath, "admin", "cluster_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rawConfig map[string]interface{}
	err = json.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster_config.json: %w", err)
	}

	var configEntries []models.ClusterConfigEntry
	for key, value := range rawConfig {
		docLink := fmt.Sprintf("https://docs.redpanda.com/current/reference/properties/cluster-properties/#%s", key)
		configEntries = append(configEntries, models.ClusterConfigEntry{
			Key:     key,
			Value:   value,
			DocLink: docLink,
		})
	}

	return configEntries, nil
}

// ParseK8sResources reads and parses all supported Kubernetes JSON files in the k8s directory.
func ParseK8sResources(bundlePath string, logsOnly bool, s store.Store) (models.K8sStore, error) {
	k8sDir := filepath.Join(bundlePath, "k8s")

	// Fallback: Check if k8s dir exists in bundlePath, if not check parent
	if _, err := os.Stat(k8sDir); os.IsNotExist(err) {
		// Try parent directory (common when bundle is extracted flat)
		parentDir := filepath.Join(bundlePath, "..", "k8s")
		if _, err := os.Stat(parentDir); err == nil {
			if !logsOnly {
				slog.Info("K8s directory not found in bundle, using sibling directory", "path", parentDir)
			}
			k8sDir = parentDir
		}
	}

	k8sStore := models.K8sStore{}
	var wg sync.WaitGroup

	// Helper to parse a list file
	parseList := func(filename string, target *[]models.K8sResource) {
		defer wg.Done()
		path := filepath.Join(k8sDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				if !logsOnly {
					slog.Warn("K8s resource file not found", "path", path)
				}
			} else {
				if !logsOnly {
					slog.Warn("Error reading K8s resource file", "path", path, "error", err)
				}
			}
			return
		}
		var list models.K8sList
		if err := json.Unmarshal(data, &list); err != nil {
			if !logsOnly {
				slog.Error("Error parsing", "filename", filename, "error", err)
			}
			return
		}
		*target = list.Items
	}

	files := []struct {
		Name   string
		Target *[]models.K8sResource
	}{
		{"pods.json", &k8sStore.Pods},
		{"services.json", &k8sStore.Services},
		{"configmaps.json", &k8sStore.ConfigMaps},
		{"endpoints.json", &k8sStore.Endpoints},
		{"events.json", &k8sStore.Events},
		{"limitranges.json", &k8sStore.LimitRanges},
		{"persistentvolumeclaims.json", &k8sStore.PersistentVolumeClaims},
		{"replicationcontrollers.json", &k8sStore.ReplicationControllers},
		{"resourcequotas.json", &k8sStore.ResourceQuotas},
		{"serviceaccounts.json", &k8sStore.ServiceAccounts},
		{"statefulsets.json", &k8sStore.StatefulSets},
		{"deployments.json", &k8sStore.Deployments},
		{"nodes.json", &k8sStore.Nodes},
	}

	wg.Add(len(files))
	for _, f := range files {
		go parseList(f.Name, f.Target)
	}

	wg.Wait()

	// Insert Events into SQLite
	if len(k8sStore.Events) > 0 {
		if err := s.BulkInsertK8sEvents(k8sStore.Events); err != nil {
			if !logsOnly {
				slog.Error("Failed to insert K8s events into store", "error", err)
			}
		}
	}

	return k8sStore, nil
}

// ParseRedpandaDataDirectory reads redpanda.yaml and extracts the data directory.
func ParseRedpandaDataDirectory(bundlePath string) (string, error) {
	filePath := filepath.Join(bundlePath, "redpanda.yaml")
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Look for data_directory: or directory:
		if strings.Contains(line, "data_directory:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				// Remove potential inline comments
				if idx := strings.Index(val, "#"); idx != -1 {
					val = strings.TrimSpace(val[:idx])
				}
				return val, nil
			}
		}
	}

	return "", fmt.Errorf("data_directory not found in redpanda.yaml")
}

// ParseRedpandaConfig reads redpanda.yaml and extracts key-value pairs.
// It performs a simple line-by-line parse to avoid external YAML dependencies.
// It handles:
// - Top level keys (redpanda:)
// - Indented keys (  data_directory: ...)
// - Inline comments
func ParseRedpandaConfig(bundlePath string) (map[string]interface{}, error) {
	filePath := filepath.Join(bundlePath, "redpanda.yaml")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	config := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	// Track context for indentation
	var parentKeys []string
	var lastIndentation int

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines or full comments
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Calculate indentation
		indentation := len(line) - len(strings.TrimLeft(line, " "))

		// Handle list items (e.g. - --flag=value)
		if strings.HasPrefix(trimmedLine, "- ") {
			// Associate with the parent key to keep the map flat but inclusive
			listKey := "list_item_" + strings.Join(parentKeys, ".") + "_" + strconv.Itoa(len(config))
			config[listKey] = strings.TrimPrefix(trimmedLine, "- ")
			continue
		}

		// Split key and value
		parts := strings.SplitN(trimmedLine, ":", 2)
		key := strings.TrimSpace(parts[0])

		if len(parts) < 2 {
			continue
		}

		value := strings.TrimSpace(parts[1])
		// Remove inline comments
		if idx := strings.Index(value, " #"); idx != -1 {
			value = strings.TrimSpace(value[:idx])
		}

		// Update context based on indentation
		if indentation == 0 {
			parentKeys = []string{}
		} else if indentation > lastIndentation {
			// Going deeper
		} else if indentation < lastIndentation {
			// Going back up - this is a rough approximation
			// A true parser would need a stack of indentations
			// For simple redpanda.yaml this often suffices to just reset if we drop back
			// But let's try to be slightly smarter: 2 spaces per level is standard
			level := indentation / 2
			if level < len(parentKeys) {
				parentKeys = parentKeys[:level]
			}
		}

		lastIndentation = indentation

		// Construct full key
		fullKey := key
		if len(parentKeys) > 0 {
			fullKey = strings.Join(parentKeys, ".") + "." + key
		}

		// If value is empty, it's likely a parent key
		if value == "" {
			parentKeys = append(parentKeys, key)
		} else {
			config[fullKey] = value
		}
	}

	return config, nil
}

// ParseMetricDefinitions extracts metric names and help text from a Prometheus format file.
func ParseMetricDefinitions(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	definitions := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Regex for HELP line: # HELP metric_name Description text
	re := regexp.MustCompile(`^# HELP ([a-zA-Z0-9_:]+) (.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# HELP") {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				definitions[matches[1]] = matches[2]
			}
		}
	}

	return definitions, scanner.Err()
}

// ParsePrometheusMetrics parses a file containing Prometheus text-format metrics
// and stores them in the provided store.
func ParsePrometheusMetrics(filePath string, allowedMetrics map[string]struct{}, s store.Store, metricTimestamp time.Time) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var batch []*models.PrometheusMetric
	const batchSize = 5000 // Increased batch size for better throughput
	
	// Add GC control variables
	var processedBytes int64
	const gcThreshold = 50 * 1024 * 1024 // 50MB

	scanner := bufio.NewScanner(file)

	// Buffer for scanner to handle potentially long lines, though metrics are usually short
	const maxCapacity = 5 * 1024 * 1024 // Increased to 5MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		
		processedBytes += int64(len(line))
		if processedBytes > gcThreshold {
			runtime.GC()
			processedBytes = 0
		}
		
		// Skip comments and empty lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Fast string parsing instead of Regex
		// Format: metric_name{label="value"} 1.23

		// 1. Find the last space which separates the value from the rest
		// We use LastIndex because the value is always at the end, but labels might contain spaces (rare but possible)
		// Actually, Prometheus format guarantees the value is the last token.
		lastSpace := strings.LastIndexByte(line, ' ')
		if lastSpace == -1 {
			continue
		}

		valueStr := line[lastSpace+1:]
		rest := line[:lastSpace] // name + labels

		// Parse value
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			continue // Skip lines with invalid values, NaN, or Inf
		}

		var name string
		var labelsJSON string

		// 2. Check for labels
		openBrace := strings.IndexByte(rest, '{')
		closeBrace := strings.LastIndexByte(rest, '}')

		if openBrace != -1 && closeBrace != -1 && closeBrace > openBrace {
			name = rest[:openBrace]
			labelsContent := rest[openBrace+1 : closeBrace]

			// Transform label content to JSON: key="value",key2="value2" -> {"key":"value","key2":"value2"}
			// Optimization: Avoid map allocation and json.Marshal by building JSON string directly
			var sb strings.Builder
			sb.WriteByte('{')
			
			// This naive split assumes no commas in values, which is generally true for Redpanda metrics
			pairs := strings.Split(labelsContent, ",")
			for i, p := range pairs {
				if i > 0 {
					sb.WriteByte(',')
				}
				eq := strings.IndexByte(p, '=')
				if eq != -1 {
					k := strings.TrimSpace(p[:eq])
					v := strings.TrimSpace(p[eq+1:]) // Keep quotes! "value"
					
					// Key needs quotes
					sb.WriteByte('"')
					sb.WriteString(k)
					sb.WriteString(`":`)
					sb.WriteString(v)
				}
			}
			sb.WriteByte('}')
			labelsJSON = sb.String()
		} else {
			name = strings.TrimSpace(rest)
			labelsJSON = "{}"
		}

		if allowedMetrics != nil {
			if _, ok := allowedMetrics[name]; !ok {
				continue
			}
		}

		batch = append(batch, &models.PrometheusMetric{
			Name:       name,
			LabelsJSON: labelsJSON,
			Value:      value,
		})

		if len(batch) >= batchSize {
			if err := s.BulkInsertMetrics(batch, metricTimestamp); err != nil {
				return fmt.Errorf("failed to bulk insert metrics: %w", err)
			}
			batch = batch[:0] // Clear batch
		}
	}

	if len(batch) > 0 { // Insert any remaining metrics
		if err := s.BulkInsertMetrics(batch, metricTimestamp); err != nil {
			return fmt.Errorf("failed to bulk insert remaining metrics: %w", err)
		}
	}

	return scanner.Err()
}
