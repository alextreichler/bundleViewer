package models

import "time"

type SarEntry struct {
	Timestamp time.Time `json:"timestamp"`
	CPUUser   float64   `json:"cpu_user"`
	CPUSystem float64   `json:"cpu_system"`
	CPUIdle   float64   `json:"cpu_idle"`
	MemUsed   float64   `json:"mem_used"` // Percent
}

type SarData struct {
	Entries []SarEntry
}
