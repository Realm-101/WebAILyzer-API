package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHTTPClientTimeout(t *testing.T) {
	// Test that our HTTP client has proper timeout configuration
	client := createHTTPClient()
	
	if client.Timeout != 15*time.Second {
		t.Errorf("Expected client timeout of 15s, got %v", client.Timeout)
	}
	
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport")
	}
	
	if transport.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("Expected ResponseHeaderTimeout of 10s, got %v", transport.ResponseHeaderTimeout)
	}
	
	if transport.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("Expected TLSHandshakeTimeout of 10s, got %v", transport.TLSHandshakeTimeout)
	}
}

func TestContextTimeout(t *testing.T) {
	// Test that context timeout works properly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	start := time.Now()
	<-ctx.Done()
	duration := time.Since(start)
	
	// Should timeout around 100ms
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Context timeout took %v, expected around 100ms", duration)
	}
	
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got %v", ctx.Err())
	}
}

func TestRedirectLimit(t *testing.T) {
	// Test that our HTTP client limits redirects
	client := createHTTPClient()
	
	// Create a request to test redirect handling
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Test the CheckRedirect function with multiple redirects
	var redirectCount int
	for i := 0; i < 15; i++ {
		err := client.CheckRedirect(req, make([]*http.Request, i))
		if err != nil {
			redirectCount = i
			break
		}
	}
	
	if redirectCount != 10 {
		t.Errorf("Expected redirect limit of 10, got %d", redirectCount)
	}
}