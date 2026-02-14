package models

import (
	"sync"
	"time"
)

type ProgressTracker struct {
	mu        sync.RWMutex
	Current   int
	Total     int
	Status    string
	Finished  bool
	StartTime time.Time
}

func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		Total:     total,
		Status:    "Initializing...",
		StartTime: time.Now(),
	}
}

func (p *ProgressTracker) Reset(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current = 0
	p.Total = total
	p.Status = "Initializing..."
	p.Finished = false
	p.StartTime = time.Now()
}

func (p *ProgressTracker) Update(increment int, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current += increment
	if status != "" {
		p.Status = status
	}
	if p.Total > 0 && p.Current > p.Total {
		p.Current = p.Total
	}
}

func (p *ProgressTracker) AddTotal(increment int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Total += increment
}

func (p *ProgressTracker) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
}

func (p *ProgressTracker) Get() (int, string, bool, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	percentage := 0
	if p.Total > 0 {
		percentage = (p.Current * 100) / p.Total
	}

	etaSeconds := -1
	if percentage > 5 && !p.Finished {
		elapsed := time.Since(p.StartTime).Seconds()
		totalEstimated := (elapsed / float64(percentage)) * 100
		remaining := totalEstimated - elapsed
		etaSeconds = int(remaining)
	}

	return percentage, p.Status, p.Finished, etaSeconds
}

func (p *ProgressTracker) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current = p.Total
	p.Finished = true
	p.Status = "Complete"
}
