package models

// DuEntry represents a single line from the du.txt output.
type DuEntry struct {
	Size int64
	Path string
}

// DiskStats represents the data from disk_stat_data/cache.json
type DiskStats struct {
	FreeBytes  int64 `json:"free_bytes"`
	TotalBytes int64 `json:"total_bytes"`
}

// ResourceUsage represents the data from resource-usage.json
type ResourceUsage struct {
	CPUPercentage float64 `json:"cpuPercentage"`
	FreeMemoryMB  float64 `json:"freeMemoryMB"`
	FreeSpaceMB   float64 `json:"freeSpaceMB"`
}

// TopicStorageStats holds storage info for a single topic
type TopicStorageStats struct {
	TopicName    string `json:"topic_name"`
	Active       bool   `json:"active"`
	SegmentCount int    `json:"segment_count"`
	SizeBytes    int64  `json:"size_bytes"`
	StaleBytes   int64  `json:"stale_bytes"`
}

// StorageAnalysis holds the aggregate results of the storage analysis
type StorageAnalysis struct {
	Topics           []TopicStorageStats
	ActiveCount      int
	OrphanedCount    int
	EffectiveDataDir string // Deduced from data-dir.txt paths
}