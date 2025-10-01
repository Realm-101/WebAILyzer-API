package wappalyzer

import (
	"sync"
	"time"
)

// PerformanceMetrics tracks performance statistics
type PerformanceMetrics struct {
	mu                sync.RWMutex
	TotalRequests     int64         `json:"total_requests"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	FastestDuration   time.Duration `json:"fastest_duration"`
	SlowestDuration   time.Duration `json:"slowest_duration"`
	TechnologiesFound int64         `json:"technologies_found"`
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		FastestDuration: time.Hour, // Initialize with a large value
	}
}

// RecordRequest records metrics for a fingerprinting request
func (pm *PerformanceMetrics) RecordRequest(duration time.Duration, techCount int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.TotalRequests++
	pm.TotalDuration += duration
	pm.TechnologiesFound += int64(techCount)
	pm.AverageDuration = pm.TotalDuration / time.Duration(pm.TotalRequests)

	if duration < pm.FastestDuration {
		pm.FastestDuration = duration
	}
	if duration > pm.SlowestDuration {
		pm.SlowestDuration = duration
	}
}

// GetMetrics returns a copy of the current metrics
func (pm *PerformanceMetrics) GetMetrics() PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return *pm
}

// Reset clears all metrics
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.TotalRequests = 0
	pm.TotalDuration = 0
	pm.AverageDuration = 0
	pm.FastestDuration = time.Hour
	pm.SlowestDuration = 0
	pm.TechnologiesFound = 0
}