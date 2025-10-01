package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	
	"github.com/webailyzer/webailyzer-lite-api/internal/database"
	"github.com/webailyzer/webailyzer-lite-api/internal/cache"
)

// MonitoringService provides system monitoring and metrics collection
type MonitoringService struct {
	collector    *MetricsCollector
	logger       *logrus.Logger
	dbConn       *database.Connection
	cacheService *cache.CacheService
	
	// Cache hit tracking
	cacheHits   int64
	cacheMisses int64
	cacheMutex  sync.RWMutex
	
	// Background monitoring
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(collector *MetricsCollector, logger *logrus.Logger, dbConn *database.Connection, cacheService *cache.CacheService) *MonitoringService {
	return &MonitoringService{
		collector:    collector,
		logger:       logger,
		dbConn:       dbConn,
		cacheService: cacheService,
		stopChan:     make(chan struct{}),
	}
}

// Start begins background monitoring tasks
func (ms *MonitoringService) Start(ctx context.Context) {
	ms.logger.Info("Starting monitoring service")
	
	// Start periodic metrics collection
	ms.wg.Add(1)
	go ms.collectSystemMetrics(ctx)
	
	// Start cache hit ratio tracking
	ms.wg.Add(1)
	go ms.trackCacheMetrics(ctx)
}

// Stop stops the monitoring service
func (ms *MonitoringService) Stop() {
	ms.logger.Info("Stopping monitoring service")
	close(ms.stopChan)
	ms.wg.Wait()
}

// collectSystemMetrics periodically collects system-level metrics
func (ms *MonitoringService) collectSystemMetrics(ctx context.Context) {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ms.stopChan:
			return
		case <-ticker.C:
			ms.updateSystemMetrics(ctx)
		}
	}
}

// updateSystemMetrics updates various system metrics
func (ms *MonitoringService) updateSystemMetrics(ctx context.Context) {
	// Update database connection metrics
	if ms.dbConn != nil && ms.dbConn.Pool != nil {
		stats := ms.dbConn.Pool.Stat()
		ms.collector.SetDatabaseConnections(int(stats.TotalConns()))
	}
	
	// Update workspace count (this would require a repository call in real implementation)
	// For now, we'll set a placeholder
	ms.collector.SetWorkspaceCount(1) // This should be fetched from database
	
	ms.logger.Debug("Updated system metrics")
}

// trackCacheMetrics tracks cache hit/miss ratios
func (ms *MonitoringService) trackCacheMetrics(ctx context.Context) {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ms.stopChan:
			return
		case <-ticker.C:
			ms.updateCacheHitRatio()
		}
	}
}

// updateCacheHitRatio calculates and updates the cache hit ratio
func (ms *MonitoringService) updateCacheHitRatio() {
	ms.cacheMutex.RLock()
	hits := ms.cacheHits
	misses := ms.cacheMisses
	ms.cacheMutex.RUnlock()
	
	total := hits + misses
	if total > 0 {
		ratio := float64(hits) / float64(total)
		ms.collector.SetCacheHitRatio(ratio)
		
		ms.logger.WithFields(logrus.Fields{
			"hits":   hits,
			"misses": misses,
			"ratio":  ratio,
		}).Debug("Updated cache hit ratio")
	}
}

// RecordCacheHit records a cache hit
func (ms *MonitoringService) RecordCacheHit() {
	ms.cacheMutex.Lock()
	ms.cacheHits++
	ms.cacheMutex.Unlock()
	
	ms.collector.RecordCacheOperation("get", "hit")
}

// RecordCacheMiss records a cache miss
func (ms *MonitoringService) RecordCacheMiss() {
	ms.cacheMutex.Lock()
	ms.cacheMisses++
	ms.cacheMutex.Unlock()
	
	ms.collector.RecordCacheOperation("get", "miss")
}

// RecordCacheSet records a cache set operation
func (ms *MonitoringService) RecordCacheSet(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	ms.collector.RecordCacheOperation("set", status)
}

// RecordCacheDelete records a cache delete operation
func (ms *MonitoringService) RecordCacheDelete(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	ms.collector.RecordCacheOperation("delete", status)
}

// GetCollector returns the metrics collector
func (ms *MonitoringService) GetCollector() *MetricsCollector {
	return ms.collector
}

// GetStats returns current monitoring statistics
func (ms *MonitoringService) GetStats() map[string]interface{} {
	ms.cacheMutex.RLock()
	hits := ms.cacheHits
	misses := ms.cacheMisses
	ms.cacheMutex.RUnlock()
	
	total := hits + misses
	ratio := 0.0
	if total > 0 {
		ratio = float64(hits) / float64(total)
	}
	
	stats := map[string]interface{}{
		"cache": map[string]interface{}{
			"hits":      hits,
			"misses":    misses,
			"hit_ratio": ratio,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	// Add database stats if available
	if ms.dbConn != nil && ms.dbConn.Pool != nil {
		dbStats := ms.dbConn.Pool.Stat()
		stats["database"] = map[string]interface{}{
			"total_connections":    dbStats.TotalConns(),
			"idle_connections":     dbStats.IdleConns(),
			"acquired_connections": dbStats.AcquiredConns(),
		}
	}
	
	return stats
}