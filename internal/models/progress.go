package models

import (
	"sync"
)

type ProgressTracker struct {
	mu       sync.RWMutex
	Current  int
	Total    int
	Status   string
	Finished bool
}

func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		Total:  total,
		Status: "Initializing...",
	}
}

func (p *ProgressTracker) Update(increment int, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current += increment
	if status != "" {
		p.Status = status
	}
	if p.Current > p.Total {
		p.Current = p.Total
	}
}

func (p *ProgressTracker) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
}

func (p *ProgressTracker) Get() (int, string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	percentage := 0
	if p.Total > 0 {
		percentage = (p.Current * 100) / p.Total
	}
	return percentage, p.Status, p.Finished
}

func (p *ProgressTracker) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current = p.Total
	p.Finished = true
	p.Status = "Complete"
}
