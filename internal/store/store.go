package store

import (
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// Store defines the interface for data storage operations.
type Store interface {
	HasLogs() (bool, error)
	InsertLog(log *models.LogEntry) error
	BulkInsertLogs(logs []*models.LogEntry) error
	GetLogs(filter *models.LogFilter) ([]*models.LogEntry, int, error)
	GetLogPatterns() ([]models.LogPattern, int, error)
	GetMinMaxLogTime() (minTime, maxTime time.Time, err error)
	GetDistinctLogNodes() ([]string, error)
	GetDistinctLogComponents() ([]string, error)
	GetDistinctMetricNames() ([]string, error)

	BulkInsertK8sEvents(events []models.K8sResource) error
	
	GetLogCountsByTime(bucketSize time.Duration) (map[time.Time]int, error)
	GetK8sEventCountsByTime(bucketSize time.Duration) (map[time.Time]int, error)

	InsertMetric(metric *models.PrometheusMetric, timestamp time.Time) error
	BulkInsertMetrics(metrics []*models.PrometheusMetric, timestamp time.Time) error
	GetMetrics(name string, labels map[string]string, startTime, endTime time.Time, limit, offset int) ([]*models.PrometheusMetric, error)

	BulkInsertCpuProfiles(profiles []models.CpuProfileEntry) error
	GetCpuProfiles() ([]models.CpuProfileAggregate, error)
	GetCpuProfileDetails(node string, shardID int, group string) ([]models.CpuProfileDetail, error)

	Optimize(p *models.ProgressTracker) error
	Close() error
}
