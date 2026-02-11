package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// AnalyzeStorage processes the data-dir.txt and cluster_partitions.json to analyze storage usage per topic.
func AnalyzeStorage(bundlePath string) (*models.StorageAnalysis, error) {
	// 1. Get Active Topics
	partitions, err := parseClusterPartitions(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster partitions: %w", err)
	}

	activeTopics := make(map[string]struct{})
	for _, p := range partitions {
		if p.Ns == "kafka" {
			activeTopics[p.Topic] = struct{}{}
		}
	}

	// 2. Parse data-dir.txt
	dataDirFile := filepath.Join(bundlePath, "data-dir.txt")
	f, err := os.Open(dataDirFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open data-dir.txt: %w", err)
	}
	defer f.Close()

	// Regex for kafka data paths: .../kafka/<topic>/<partition>_<revision>/<segment>.log
	// Example: /var/lib/redpanda/data/kafka/test-topic/0_15/0-1-v1.log
	// Captures: 1=Topic, 2=PartitionID, 3=RevisionID
	kafkaPathRe := regexp.MustCompile(`/kafka/([^/]+)/(\d+)_(\d+)/[^/]+\.log$`)

	topicStats := make(map[string]*models.TopicStorageStats)
	var effectiveDataDir string
	
	// Map "Topic/PartitionID" -> Active RevisionID from cluster state
	activeRevisions := make(map[string]int)
	for _, p := range partitions {
		if p.Ns == "kafka" {
			key := fmt.Sprintf("%s/%d", p.Topic, p.PartitionID)
			activeRevisions[key] = p.Revision
			// Also mark topic as generally active
			activeTopics[p.Topic] = struct{}{}
		}
	}
	
	// Use json.Decoder for streaming parse
	dec := json.NewDecoder(f)
	
	// Expect start of object
	if token, err := dec.Token(); err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("expected JSON object in data-dir.txt")
	}

	for dec.More() {
		// Read key (path)
		token, err := dec.Token()
		if err != nil {
			return nil, err
		}
		path, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key in data-dir.txt")
		}

		// Read value (info object)
		var info struct {
			SizeBytes int64 `json:"size_bytes"`
		}
		if err := dec.Decode(&info); err != nil {
			return nil, err
		}

		// Process the entry
		match := kafkaPathRe.FindStringSubmatch(path)
		if match != nil {
			if effectiveDataDir == "" {
				if idx := strings.Index(path, "/kafka/"); idx != -1 {
					effectiveDataDir = path[:idx]
				}
			}

			topic := match[1]
			partitionID := match[2] // string
			revisionIDStr := match[3] // string
			
			// Check if active
			key := fmt.Sprintf("%s/%s", topic, partitionID)
			activeRev, isActive := activeRevisions[key]
			
			// It's strictly active if topic exists AND revision matches
			// If topic matches but revision doesn't, it's an orphaned revision (stale data)
			isStale := false
			if isActive {
				currentRevStr := fmt.Sprintf("%d", activeRev)
				if revisionIDStr != currentRevStr {
					isStale = true
				}
			}

			if _, exists := topicStats[topic]; !exists {
				// Initialize
				_, topicExists := activeTopics[topic]
				topicStats[topic] = &models.TopicStorageStats{
					TopicName: topic,
					Active:    topicExists, // Generally active
				}
			}
			
			stats := topicStats[topic]
			stats.SegmentCount++
			stats.SizeBytes += info.SizeBytes
			
			if isStale {
				stats.StaleBytes += info.SizeBytes
			} else if !stats.Active {
				// Entire topic is orphaned
				stats.StaleBytes += info.SizeBytes
			}
		}
	}

	// 3. Aggregate Results
	var analysis models.StorageAnalysis
	analysis.EffectiveDataDir = effectiveDataDir

	for _, stats := range topicStats {
		if stats.Active {
			analysis.ActiveCount++
		} else {
			analysis.OrphanedCount++
		}
		// If significant stale bytes, flagging might be useful in UI
		analysis.Topics = append(analysis.Topics, *stats)
	}

	// Sort by size descending
	sort.Slice(analysis.Topics, func(i, j int) bool {
		return analysis.Topics[i].SizeBytes > analysis.Topics[j].SizeBytes
	})

	return &analysis, nil
}

// parseClusterPartitions reads and parses the cluster_partitions.json file.
func parseClusterPartitions(bundlePath string) ([]models.ClusterPartition, error) {
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
