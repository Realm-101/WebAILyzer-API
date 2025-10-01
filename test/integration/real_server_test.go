package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealServerIntegration tests against the actual compiled server
func TestRealServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real server tests in short mode")
	}

	// Build the server binary
	serverPath := buildServer(t)
	defer os.Remove(serverPath)

	// Start the server
	serverURL := startServer(t, serverPath)
	defer stopServer(t)

	// Wait for server to be ready
	waitForServer(t, serverURL)

	t.Run("Real Health Check", func(t *testing.T) {
		resp, err := http.Get(serverURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "ok", response["status"])
		assert.Contains(t, response, "memory")

		// Verify memory stats structure
		memory := response["memory"].(map[string]interface{})
		assert.Contains(t, memory, "alloc_mb")
		assert.Contains(t, memory, "total_alloc_mb")
		assert.Contains(t, memory, "sys_mb")
		assert.Contains(t, memory, "num_gc")
		assert.Contains(t, memory, "num_goroutine")
	})

	t.Run("Real URL Analysis - httpbin.org", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "https://httpbin.org/get",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "https://httpbin.org/get", response["url"])
		assert.Contains(t, response, "detected")
		assert.Contains(t, response, "content_type")

		// Verify detected technologies structure
		detected := response["detected"].(map[string]interface{})
		// httpbin.org should have some detectable technologies
		t.Logf("Detected technologies: %+v", detected)
	})

	t.Run("Real URL Analysis - example.com", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "https://example.com",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "https://example.com", response["url"])
		assert.Contains(t, response, "detected")
		assert.Contains(t, response, "content_type")

		t.Logf("Example.com analysis result: %+v", response)
	})

	t.Run("Invalid URL Handling", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "not-a-valid-url",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "type")
		assert.Equal(t, "validation_error", response["type"])
	})

	t.Run("Unreachable URL Handling", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "https://unreachable-domain-12345.example",
		}

		jsonBody, _ := json.Marshal(requestBody)
		
		client := &http.Client{
			Timeout: 30 * time.Second, // Give it time to fail properly
		}
		
		resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return a network error
		assert.Contains(t, []int{http.StatusBadGateway, http.StatusGatewayTimeout}, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "type")
		assert.Contains(t, []string{"network_error", "timeout_error"}, response["type"])
	})

	t.Run("404 URL Handling", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "https://httpbin.org/status/404",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "type")
		assert.Equal(t, "not_found_error", response["type"])
	})

	t.Run("Large Response Handling", func(t *testing.T) {
		// Test with a 1MB response
		requestBody := map[string]string{
			"url": "https://httpbin.org/bytes/1048576",
		}

		jsonBody, _ := json.Marshal(requestBody)
		
		client := &http.Client{
			Timeout: 60 * time.Second, // Give it time to download
		}
		
		resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle large responses (within 5MB limit)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "https://httpbin.org/bytes/1048576", response["url"])
		assert.Contains(t, response, "detected")
	})

	t.Run("Request Headers and User Agent", func(t *testing.T) {
		// Use httpbin.org to inspect what headers we're sending
		requestBody := map[string]string{
			"url": "https://httpbin.org/headers",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "https://httpbin.org/headers", response["url"])
		t.Logf("Headers endpoint analysis: %+v", response)
	})

	t.Run("CORS Headers", func(t *testing.T) {
		// Test OPTIONS request for CORS
		req, err := http.NewRequest("OPTIONS", serverURL+"/v1/analyze", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should have CORS headers
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Origin"), "*")
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	})
}

var serverCmd *exec.Cmd

// buildServer compiles the server binary for testing
func buildServer(t *testing.T) string {
	serverPath := "./webailyzer-api-test"
	
	cmd := exec.Command("go", "build", "-o", serverPath, "./cmd/webailyzer-api")
	cmd.Dir = "../.." // Go up to project root
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build server: %v\nOutput: %s", err, output)
	}
	
	return "../../" + serverPath
}

// startServer starts the compiled server
func startServer(t *testing.T, serverPath string) string {
	port := "18080" // Use different port to avoid conflicts
	
	serverCmd = exec.Command(serverPath)
	serverCmd.Env = append(os.Environ(), "PORT="+port)
	
	// Capture output for debugging
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	
	err := serverCmd.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	
	return "http://localhost:" + port
}

// stopServer stops the running server
func stopServer(t *testing.T) {
	if serverCmd != nil && serverCmd.Process != nil {
		// Send SIGTERM for graceful shutdown
		err := serverCmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			t.Logf("Failed to send SIGTERM: %v", err)
			// Force kill if graceful shutdown fails
			serverCmd.Process.Kill()
		}
		
		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- serverCmd.Wait()
		}()
		
		select {
		case <-done:
			// Process exited
		case <-time.After(10 * time.Second):
			// Force kill after timeout
			t.Log("Server didn't shutdown gracefully, force killing")
			serverCmd.Process.Kill()
			<-done
		}
	}
}

// waitForServer waits for the server to be ready
func waitForServer(t *testing.T, serverURL string) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		resp, err := client.Get(serverURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return // Server is ready
			}
		}
		
		time.Sleep(1 * time.Second)
	}
	
	t.Fatal("Server failed to start within 30 seconds")
}