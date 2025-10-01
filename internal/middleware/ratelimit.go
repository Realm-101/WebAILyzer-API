package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/projectdiscovery/wappalyzergo/internal/config"
)

// RateLimitEntry represents a rate limit entry for a workspace
type RateLimitEntry struct {
	Count      int
	WindowStart time.Time
	Limit      int
}

// RateLimitMiddleware provides rate limiting per workspace
type RateLimitMiddleware struct {
	entries         map[uuid.UUID]*RateLimitEntry
	mutex           sync.RWMutex
	windowDuration  time.Duration
	cleanupInterval time.Duration
	defaultLimit    int
	logger          *logrus.Logger
	stopCleanup     chan struct{}
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(cfg *config.RateLimitConfig, logger *logrus.Logger) *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		entries:         make(map[uuid.UUID]*RateLimitEntry),
		windowDuration:  cfg.WindowDuration,
		cleanupInterval: cfg.CleanupInterval,
		defaultLimit:    cfg.DefaultLimit,
		logger:          logger,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredEntries()

	return rl
}

// RateLimit applies rate limiting based on workspace
func (rl *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get auth context
		authContext, ok := GetAuthContext(r.Context())
		if !ok {
			// If no auth context, skip rate limiting (should be handled by auth middleware)
			next.ServeHTTP(w, r)
			return
		}

		// Check rate limit
		allowed, remaining, resetTime := rl.checkRateLimit(authContext.WorkspaceID, authContext.RateLimit)
		
		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(authContext.RateLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			rl.logger.WithFields(logrus.Fields{
				"workspace_id": authContext.WorkspaceID,
				"limit":        authContext.RateLimit,
			}).Warn("Rate limit exceeded")

			rl.writeRateLimitErrorResponse(w, authContext.RateLimit, resetTime)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkRateLimit checks if the request is within rate limits
func (rl *RateLimitMiddleware) checkRateLimit(workspaceID uuid.UUID, limit int) (allowed bool, remaining int, resetTime time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Use default limit if workspace limit is 0
	if limit <= 0 {
		limit = rl.defaultLimit
	}

	entry, exists := rl.entries[workspaceID]
	if !exists {
		// First request for this workspace
		rl.entries[workspaceID] = &RateLimitEntry{
			Count:       1,
			WindowStart: now,
			Limit:       limit,
		}
		return true, limit - 1, now.Add(rl.windowDuration)
	}

	// Check if we need to reset the window
	if now.Sub(entry.WindowStart) >= rl.windowDuration {
		entry.Count = 1
		entry.WindowStart = now
		entry.Limit = limit
		return true, limit - 1, now.Add(rl.windowDuration)
	}

	// Check if within limit
	if entry.Count >= limit {
		resetTime = entry.WindowStart.Add(rl.windowDuration)
		return false, 0, resetTime
	}

	// Increment count
	entry.Count++
	entry.Limit = limit // Update limit in case it changed
	resetTime = entry.WindowStart.Add(rl.windowDuration)
	return true, limit - entry.Count, resetTime
}

// cleanupExpiredEntries removes expired rate limit entries
func (rl *RateLimitMiddleware) cleanupExpiredEntries() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes expired entries
func (rl *RateLimitMiddleware) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for workspaceID, entry := range rl.entries {
		if now.Sub(entry.WindowStart) >= rl.windowDuration {
			delete(rl.entries, workspaceID)
		}
	}

	rl.logger.WithField("remaining_entries", len(rl.entries)).Debug("Rate limit cleanup completed")
}

// Stop stops the cleanup goroutine
func (rl *RateLimitMiddleware) Stop() {
	close(rl.stopCleanup)
}

// writeRateLimitErrorResponse writes a rate limit exceeded error response
func (rl *RateLimitMiddleware) writeRateLimitErrorResponse(w http.ResponseWriter, limit int, resetTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
	w.WriteHeader(http.StatusTooManyRequests)

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "RATE_LIMIT_EXCEEDED",
			"message": fmt.Sprintf("Rate limit of %d requests per hour exceeded", limit),
			"details": map[string]interface{}{
				"limit":      limit,
				"reset_time": resetTime.Unix(),
				"retry_after": int64(time.Until(resetTime).Seconds()),
			},
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// GetRateLimitStats returns current rate limit statistics for debugging
func (rl *RateLimitMiddleware) GetRateLimitStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_workspaces": len(rl.entries),
		"window_duration":  rl.windowDuration.String(),
		"default_limit":    rl.defaultLimit,
	}

	workspaceStats := make(map[string]interface{})
	for workspaceID, entry := range rl.entries {
		workspaceStats[workspaceID.String()] = map[string]interface{}{
			"count":        entry.Count,
			"limit":        entry.Limit,
			"window_start": entry.WindowStart.Unix(),
			"remaining":    entry.Limit - entry.Count,
		}
	}
	stats["workspaces"] = workspaceStats

	return stats
}