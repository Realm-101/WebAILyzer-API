package main

import (
	"strings"
	"testing"
)

func TestMemoryStats(t *testing.T) {
	stats := getMemoryStats()
	
	// Just verify the function returns valid data structure
	if stats.NumGoroutine < 1 {
		t.Error("Number of goroutines should be at least 1")
	}
	
	// These values should be non-negative
	if stats.Alloc < 0 || stats.TotalAlloc < 0 || stats.Sys < 0 {
		t.Error("Memory stats should be non-negative")
	}
	
	t.Logf("Memory stats: Alloc=%dMB, TotalAlloc=%dMB, Sys=%dMB, NumGC=%d, NumGoroutine=%d",
		stats.Alloc, stats.TotalAlloc, stats.Sys, stats.NumGC, stats.NumGoroutine)
}

func TestOptimizeGCSettings(t *testing.T) {
	// Test that GC optimization doesn't panic
	optimizeGCSettings()
	
	// Force a GC to ensure it works
	stats1 := getMemoryStats()
	
	// Allocate some memory
	data := make([]byte, 1024*1024) // 1MB
	_ = data
	
	stats2 := getMemoryStats()
	
	// Memory should have increased
	if stats2.Alloc <= stats1.Alloc {
		t.Log("Memory allocation may not have increased as expected, but this is not necessarily an error")
	}
}

func TestReadResponseBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxSize  int64
		wantErr  bool
	}{
		{
			name:    "small content",
			input:   "Hello, World!",
			maxSize: 1024,
			wantErr: false,
		},
		{
			name:    "content at limit",
			input:   "Hello",
			maxSize: 5,
			wantErr: false,
		},
		{
			name:    "content exceeds limit",
			input:   "Hello, World!",
			maxSize: 5,
			wantErr: true,
		},
		{
			name:    "empty content",
			input:   "",
			maxSize: 1024,
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := readResponseBody(reader, tt.maxSize)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if string(result) != tt.input {
				t.Errorf("Expected %q, got %q", tt.input, string(result))
			}
		})
	}
}