package models

import "time"

// PrometheusMetric represents a single metric sample
type PrometheusMetric struct {
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricsBundle holds parsed metrics from a file or set of files
type MetricsBundle struct {
	// Map of filename -> metrics
	Files map[string][]PrometheusMetric
}
