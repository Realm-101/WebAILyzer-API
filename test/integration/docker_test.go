package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerIntegration tests the application running in a Docker container
func TestDockerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	// Check if Docker is available
	if !isDockerAvailable(t) {
		t.Skip("Docker is not available, skipping Docker integration tests")
	}

	// Build Docker image
	imageName := buildDockerImage(t)
	defer cleanupDockerImage(t, imageName)

	// Start container
	containerID, containerURL := startDockerContainer(t, imageName)
	defer stopDockerContainer(t, containerID)

	// Wait for container to be ready
	waitForDockerContainer(t, containerURL)

	t.Run("Docker Health Check", func(t *testing.T) {
		resp, err := http.Get(containerURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "ok", response["status"])
		assert.Contains(t, response, "memory")
	})

	t.Run("Docker URL Analysis", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "https://httpbin.org/get",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(containerURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "https://httpbin.org/get", response["url"])
		assert.Contains(t, response, "detected")
		assert.Contains(t, response, "content_type")
	})

	t.Run("Docker Error Handling", func(t *testing.T) {
		requestBody := map[string]string{
			"url": "invalid-url",
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(containerURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "type")
	})

	t.Run("Docker Container Resource Limits", func(t *testing.T) {
		// Test multiple concurrent requests to verify resource handling
		const numRequests = 5
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				requestBody := map[string]string{
					"url": fmt.Sprintf("https://httpbin.org/delay/1"),
				}

				jsonBody, _ := json.Marshal(requestBody)
				
				client := &http.Client{
					Timeout: 30 * time.Second,
				}
				
				resp, err := client.Post(containerURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
				if err != nil {
					results <- 0
					return
				}
				defer resp.Body.Close()
				results <- resp.StatusCode
			}(i)
		}

		successCount := 0
		for i := 0; i < numRequests; i++ {
			select {
			case code := <-results:
				if code == http.StatusOK {
					successCount++
				}
			case <-time.After(60 * time.Second):
				t.Fatal("Timeout waiting for concurrent Docker requests")
			}
		}

		assert.GreaterOrEqual(t, successCount, numRequests/2, "At least half of concurrent requests should succeed in Docker")
	})

	t.Run("Docker Health Check Endpoint", func(t *testing.T) {
		// Test the Docker health check mechanism
		cmd := exec.Command("docker", "exec", containerID, "curl", "-f", "http://localhost:8080/health")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Health check command output: %s", output)
			t.Logf("Health check command error: %v", err)
		}
		
		// The health check should succeed (exit code 0)
		assert.NoError(t, err, "Docker health check should succeed")
		
		// Verify the output contains expected JSON
		assert.Contains(t, string(output), "\"status\":\"ok\"")
	})

	t.Run("Docker Container Logs", func(t *testing.T) {
		// Check container logs for any errors
		cmd := exec.Command("docker", "logs", containerID)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		logs := string(output)
		t.Logf("Container logs:\n%s", logs)

		// Should contain startup message
		assert.Contains(t, logs, "Starting WebAIlyzer Lite API server on port 8080")
		
		// Should not contain fatal errors
		assert.NotContains(t, logs, "FATAL")
		assert.NotContains(t, logs, "panic")
	})
}

// isDockerAvailable checks if Docker is available on the system
func isDockerAvailable(t *testing.T) bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// buildDockerImage builds the Docker image for testing
func buildDockerImage(t *testing.T) string {
	imageName := "webailyzer-lite-api:test"
	
	t.Logf("Building Docker image: %s", imageName)
	
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = "../.." // Go up to project root
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build Docker image: %v\nOutput: %s", err, output)
	}
	
	t.Logf("Docker image built successfully")
	return imageName
}

// startDockerContainer starts a Docker container and returns container ID and URL
func startDockerContainer(t *testing.T, imageName string) (string, string) {
	port := "18081" // Use different port to avoid conflicts
	
	t.Logf("Starting Docker container on port %s", port)
	
	cmd := exec.Command("docker", "run", "-d", "-p", port+":8080", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start Docker container: %v\nOutput: %s", err, output)
	}
	
	containerID := strings.TrimSpace(string(output))
	containerURL := "http://localhost:" + port
	
	t.Logf("Docker container started: %s", containerID[:12])
	return containerID, containerURL
}

// stopDockerContainer stops and removes the Docker container
func stopDockerContainer(t *testing.T, containerID string) {
	t.Logf("Stopping Docker container: %s", containerID[:12])
	
	// Stop the container
	cmd := exec.Command("docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to stop Docker container: %v\nOutput: %s", err, output)
	}
	
	// Remove the container
	cmd = exec.Command("docker", "rm", containerID)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to remove Docker container: %v\nOutput: %s", err, output)
	}
}

// cleanupDockerImage removes the Docker image
func cleanupDockerImage(t *testing.T, imageName string) {
	t.Logf("Cleaning up Docker image: %s", imageName)
	
	cmd := exec.Command("docker", "rmi", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to remove Docker image: %v\nOutput: %s", err, output)
	}
}

// waitForDockerContainer waits for the Docker container to be ready
func waitForDockerContainer(t *testing.T, containerURL string) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	
	t.Log("Waiting for Docker container to be ready...")
	
	for i := 0; i < 60; i++ { // Wait up to 60 seconds
		resp, err := client.Get(containerURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Log("Docker container is ready")
				return
			}
		}
		
		if i%10 == 0 {
			t.Logf("Still waiting for container... (%d/60)", i)
		}
		
		time.Sleep(1 * time.Second)
	}
	
	t.Fatal("Docker container failed to start within 60 seconds")
}