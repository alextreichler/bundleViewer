package models

type CpuProfile struct {
	Schema         int            `json:"schema"`
	Arch           string         `json:"arch"`
	Version        string         `json:"version"`
	WaitMs         int            `json:"wait_ms"`
	SamplePeriodMs int            `json:"sample_period_ms"`
	Profile        []CpuShardProfile `json:"profile"`
}

type CpuShardProfile struct {
	ShardID        int                `json:"shard_id"`
	DroppedSamples int                `json:"dropped_samples"`
	Samples        []CpuSample `json:"samples"`
}

type CpuSample struct {
	UserBacktrace   string `json:"user_backtrace"`
	SchedulingGroup string `json:"scheduling_group"`
	Group           string `json:"group"` // Some versions use 'group'
	Occurrences     int    `json:"occurrences"`
}

// Flat structure for DB
type CpuProfileEntry struct {
	Node            string
	ShardID         int
	SchedulingGroup string
	UserBacktrace   string
	Occurrences     int
}

// Aggregated view for UI
type CpuProfileAggregate struct {
	Node            string  `json:"node"`
	ShardID         int     `json:"shard_id"`
	SchedulingGroup string  `json:"scheduling_group"`
	TotalSamples    int     `json:"total_samples"`
	Percentage      float64 `json:"percentage"`
}

type CpuProfileDetail struct {
	UserBacktrace string  `json:"user_backtrace"`
	Occurrences   int     `json:"occurrences"`
	Percentage    float64 `json:"percentage"`
}
