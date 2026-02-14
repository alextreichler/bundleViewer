package models

import "time"

type LogEntry struct {
	ID         int64     `json:"id"`
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
	IsPinned    bool        `json:"isPinned"`
}

// Reset clears the LogEntry for reuse in a sync.Pool
func (le *LogEntry) Reset() {
	le.ID = 0
	le.Timestamp = time.Time{}
	le.Level = ""
	le.Node = ""
	le.Shard = ""
	le.Component = ""
	le.Message = ""
	le.Raw = ""
	le.LineNumber = 0
	le.FilePath = ""
	le.Fingerprint = ""
	le.Snippet = ""
	le.Insight = nil
	le.IsPinned = false
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

type PinnedLog struct {
	LogEntry
	Note string `json:"note"`
}
