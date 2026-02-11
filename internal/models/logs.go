package models

import "time"

type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Node       string    `json:"node"`
	Shard      string    `json:"shard"`
	Component  string    `json:"component"`
	Message    string    `json:"message"`
	Raw        string    `json:"raw"` // Keep raw for full context if needed
	LineNumber int       `json:"lineNumber"`
	FilePath   string    `json:"filePath"`
	Fingerprint string   `json:"fingerprint"`
	Snippet     string   `json:"snippet"` // FTS snippet with highlighting
	Insight     *LogInsight `json:"insight,omitempty"`
}

type LogInsight struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Action      string `json:"action"`
}

type LogFilter struct {
	Level     []string
	Node      []string
	Component []string
	Search    string
	Ignore    string
	StartTime string
	EndTime   string
	Sort      string // "asc" or "desc"
	Limit     int
	Offset    int
}

type LogPattern struct {
	Signature   string      `json:"signature"`
	Count       int         `json:"count"`
	Level       string      `json:"level"`
	SampleEntry LogEntry    `json:"sample_entry"`
	FirstSeen   time.Time   `json:"first_seen"`
	LastSeen    time.Time   `json:"last_seen"`
	Insight     *LogInsight `json:"insight,omitempty"`
}
