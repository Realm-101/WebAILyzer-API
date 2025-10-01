package wappalyzer

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached fingerprint result
type CacheEntry struct {
	Result    map[string]struct{}
	Timestamp time.Time
	TTL       time.Duration
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Since(ce.Timestamp) > ce.TTL
}

// FingerprintCache provides caching for fingerprint results
type FingerprintCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
}

// NewFingerprintCache creates a new fingerprint cache
func NewFingerprintCache(maxSize int, ttl time.Duration) *FingerprintCache {
	return &FingerprintCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// generateKey creates a cache key from headers and body
func (fc *FingerprintCache) generateKey(headers map[string][]string, body []byte) string {
	h := md5.New()
	
	// Hash headers
	for key, values := range headers {
		h.Write([]byte(key))
		for _, value := range values {
			h.Write([]byte(value))
		}
	}
	
	// Hash body (first 1KB for performance)
	bodyToHash := body
	if len(body) > 1024 {
		bodyToHash = body[:1024]
	}
	h.Write(bodyToHash)
	
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get retrieves a cached result
func (fc *FingerprintCache) Get(headers map[string][]string, body []byte) (map[string]struct{}, bool) {
	key := fc.generateKey(headers, body)
	
	fc.mu.RLock()
	entry, exists := fc.entries[key]
	fc.mu.RUnlock()
	
	if !exists || entry.IsExpired() {
		if exists {
			// Clean up expired entry
			fc.mu.Lock()
			delete(fc.entries, key)
			fc.mu.Unlock()
		}
		return nil, false
	}
	
	return entry.Result, true
}

// Set stores a result in the cache
func (fc *FingerprintCache) Set(headers map[string][]string, body []byte, result map[string]struct{}) {
	key := fc.generateKey(headers, body)
	
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	// Check if we need to evict entries
	if len(fc.entries) >= fc.maxSize {
		fc.evictOldest()
	}
	
	fc.entries[key] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		TTL:       fc.ttl,
	}
}

// evictOldest removes the oldest cache entry
func (fc *FingerprintCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range fc.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}
	
	if oldestKey != "" {
		delete(fc.entries, oldestKey)
	}
}

// Clear removes all cache entries
func (fc *FingerprintCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.entries = make(map[string]*CacheEntry)
}

// Size returns the current cache size
func (fc *FingerprintCache) Size() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.entries)
}

// Stats returns cache statistics
func (fc *FingerprintCache) Stats() map[string]interface{} {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	
	expired := 0
	for _, entry := range fc.entries {
		if entry.IsExpired() {
			expired++
		}
	}
	
	return map[string]interface{}{
		"total_entries":   len(fc.entries),
		"expired_entries": expired,
		"max_size":        fc.maxSize,
		"ttl_seconds":     fc.ttl.Seconds(),
	}
}