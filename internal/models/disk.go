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
