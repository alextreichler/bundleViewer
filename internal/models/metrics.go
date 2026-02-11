package models

import "time"

// PrometheusMetric represents a single metric sample
type PrometheusMetric struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	LabelsJSON string            `json:"-"` // Optimization: Pre-marshaled labels
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
}

// MetricsBundle holds parsed metrics from a file or set of files
type MetricsBundle struct {
	Files           map[string][]PrometheusMetric
	PerTopicMetrics map[string]float64 // topic_name -> total_value (e.g. max bytes/sec or total bytes)
}
