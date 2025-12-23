package timeline

import (
	"sort"
	"strconv"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// EventType defines the source/type of the event
type EventType string

const (
	EventTypeLog     EventType = "Log"
	EventTypeK8s     EventType = "K8s"
	EventTypeRestart EventType = "Restart"
)

// TimelineEvent represents a single point on the timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        EventType `json:"type"`
	Source      string    `json:"source"` // e.g. "Node 1", "Pod redpanda-0"
	Description string    `json:"description"`
	Level       string    `json:"level"` // "Normal", "Warning", "Critical"
	Details     string    `json:"details,omitempty"`
}

// GenerateTimeline creates a merged timeline from various sources
func GenerateTimeline(data *models.TimelineSourceData) []TimelineEvent {
	var events []TimelineEvent

	// 1. Process Logs (Only Critical/Error/Warning)
	for _, log := range data.Logs {
		if log.Level == "ERROR" || log.Level == "WARN" {
			severity := "Warning"
			if log.Level == "ERROR" {
				severity = "Critical"
			}

			events = append(events, TimelineEvent{
				Timestamp:   log.Timestamp,
				Type:        EventTypeLog,
				Source:      log.Node,
				Description: log.Message,
				Level:       severity,
				Details:     log.Raw,
			})
		}
	}

	// 2. Process K8s Events (Resource Lifecycle & System Events)
	for _, event := range data.K8sEvents {
		level := "Normal"
		if event.Type == "Warning" {
			level = "Warning"
		}

		events = append(events, TimelineEvent{
			Timestamp:   event.LastTimestamp,
			Type:        EventTypeK8s,
			Source:      event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
			Description: event.Reason + ": " + event.Message,
			Level:       level,
			Details:     "Source: " + event.Source.Component + "\nCount: " + strconv.Itoa(event.Count),
		})
	}

	// 3. Process Pod Lifecycle (Derived from Pod Status)
	for _, pod := range data.K8sPods {
		for _, status := range pod.Status.ContainerStatuses {
			// Current Running State
			if running, ok := status.State["running"]; ok && !running.StartedAt.IsZero() {
				events = append(events, TimelineEvent{
					Timestamp:   running.StartedAt,
					Type:        EventTypeK8s,
					Source:      pod.Metadata.Name,
					Description: "Pod Started (Running)",
					Level:       "Normal",
					Details:     "Container: " + status.Name,
				})
			}

			// Last Termination State (Crash/Restart)
			if terminated, ok := status.LastState["terminated"]; ok && !terminated.FinishedAt.IsZero() {
				level := "Warning"
				if terminated.Reason == "OOMKilled" || terminated.Reason == "Error" || terminated.ExitCode != 0 {
					level = "Critical"
				}

				events = append(events, TimelineEvent{
					Timestamp:   terminated.FinishedAt,
					Type:        EventTypeRestart,
					Source:      pod.Metadata.Name,
					Description: "Pod Terminated: " + terminated.Reason,
					Level:       level,
					Details:     "Exit Code: " + strconv.Itoa(terminated.ExitCode) + "\nContainer: " + status.Name,
				})
			}
		}
	}

	// Sort by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return events
}
