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

// TestEdgeCases tests various edge cases and unusual scenarios
func TestEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping edge case tests in short mode")
	}

	// Use the real server for edge case testing
	serverPath := buildServer(t)
	defer func() {
		// Clean up server binary
		exec.Command("rm", "-f", serverPath).Run()
	}()

	serverURL := startServer(t, serverPath)
	defer stopServer(t)

	waitForServer(t, serverURL)

	t.Run("Very Long URL", func(t *testing.T) {
		// Create a very long URL (but still valid)
		longPath := strings.Repeat("a", 2000)
		longURL := "https://httpbin.org/anything/" + longPath

		requestBody := map[string]string{
			"url": longURL,
		}

		jsonBody, _ := json.Marshal(requestBody)
		
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		
		resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle long URLs appropriately
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusRequestURITooLong}, resp.StatusCode)
	})

	t.Run("URL with Special Characters", func(t *testing.T) {
		// Test URL with various special characters
		specialURL := "https://httpbin.org/anything/test%20with%20spaces?param=value&other=test%2Bplus"

		requestBody := map[string]string{
			"url": specialURL,
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, specialURL, response["url"])
	})

	t.Run("International Domain Names", func(t *testing.T) {
		// Test with internationalized domain name
		idnURL := "https://xn--nxasmq6b.xn--j6w193g" // 中国.香港 in punycode

		requestBody := map[string]string{
			"url": idnURL,
		}

		jsonBody, _ := json.Marshal(requestBody)
		
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		
		resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle IDN appropriately (may fail due to network, but shouldn't crash)
		assert.Contains(t, []int{http.StatusOK, http.StatusBadGateway, http.StatusGatewayTimeout}, resp.StatusCode)
	})

	t.Run("Different HTTP Methods on Analyze Endpoint", func(t *testing.T) {
		methods := []string{"GET", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req, err := http.NewRequest(method, serverURL+"/v1/analyze", strings.NewReader("{}"))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
			})
		}
	})

	t.Run("Different Content Types", func(t *testing.T) {
		requestBody := `{"url": "https://httpbin.org/get"}`

		contentTypes := []string{
			"text/plain",
			"application/xml",
			"multipart/form-data",
			"", // No content type
		}

		for _, contentType := range contentTypes {
			t.Run("ContentType_"+contentType, func(t *testing.T) {
				req, err := http.NewRequest("POST", serverURL+"/v1/analyze", strings.NewReader(requestBody))
				require.NoError(t, err)
				
				if contentType != "" {
					req.Header.Set("Content-Type", contentType)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				// Should handle different content types appropriately
				if contentType == "application/json" || contentType == "" {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				} else {
					// May accept or reject based on implementation
					assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusUnsupportedMediaType}, resp.StatusCode)
				}
			})
		}
	})

	t.Run("Large JSON Payload", func(t *testing.T) {
		// Create a large JSON payload
		largeData := strings.Repeat("x", 10000)
		requestBody := map[string]string{
			"url":  "https://httpbin.org/get",
			"data": largeData,
		}

		jsonBody, _ := json.Marshal(requestBody)
		resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle large payloads appropriately
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusRequestEntityTooLarge}, resp.StatusCode)
	})

	t.Run("Malformed JSON Payloads", func(t *testing.T) {
		malformedPayloads := []string{
			`{"url": "https://httpbin.org/get"`, // Missing closing brace
			`{"url": https://httpbin.org/get}`,  // Missing quotes
			`{url: "https://httpbin.org/get"}`,  // Missing quotes on key
			`{"url": "https://httpbin.org/get",}`, // Trailing comma
			``,                                   // Empty payload
			`null`,                              // Null payload
			`[]`,                                // Array instead of object
			`"string"`,                          // String instead of object
		}

		for i, payload := range malformedPayloads {
			t.Run(fmt.Sprintf("MalformedJSON_%d", i), func(t *testing.T) {
				resp, err := http.Post(serverURL+"/v1/analyze", "application/json", strings.NewReader(payload))
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Contains(t, response, "error")
				assert.Contains(t, response, "type")
			})
		}
	})

	t.Run("Various URL Schemes", func(t *testing.T) {
		urlSchemes := []struct {
			url          string
			expectStatus int
		}{
			{"http://httpbin.org/get", http.StatusOK},
			{"https://httpbin.org/get", http.StatusOK},
			{"ftp://example.com", http.StatusBadRequest},
			{"file:///etc/passwd", http.StatusBadRequest},
			{"javascript:alert('xss')", http.StatusBadRequest},
			{"data:text/html,<h1>test</h1>", http.StatusBadRequest},
		}

		for _, test := range urlSchemes {
			t.Run("Scheme_"+test.url, func(t *testing.T) {
				requestBody := map[string]string{
					"url": test.url,
				}

				jsonBody, _ := json.Marshal(requestBody)
				resp, err := http.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, test.expectStatus, resp.StatusCode)
			})
		}
	})

	t.Run("Redirect Handling", func(t *testing.T) {
		// Test various redirect scenarios
		redirectTests := []struct {
			name string
			url  string
		}{
			{"Single Redirect", "https://httpbin.org/redirect/1"},
			{"Multiple Redirects", "https://httpbin.org/redirect/5"},
			{"Too Many Redirects", "https://httpbin.org/redirect/15"}, // Should hit redirect limit
		}

		for _, test := range redirectTests {
			t.Run(test.name, func(t *testing.T) {
				requestBody := map[string]string{
					"url": test.url,
				}

				jsonBody, _ := json.Marshal(requestBody)
				
				client := &http.Client{
					Timeout: 30 * time.Second,
				}
				
				resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				defer resp.Body.Close()

				// Should handle redirects appropriately
				if strings.Contains(test.name, "Too Many") {
					// May fail due to redirect limit
					assert.Contains(t, []int{http.StatusOK, http.StatusBadGateway}, resp.StatusCode)
				} else {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}
			})
		}
	})

	t.Run("Different Response Content Types", func(t *testing.T) {
		contentTypeTests := []struct {
			name string
			url  string
		}{
			{"HTML", "https://httpbin.org/html"},
			{"JSON", "https://httpbin.org/json"},
			{"XML", "https://httpbin.org/xml"},
			{"Plain Text", "https://httpbin.org/robots.txt"},
			{"Binary Data", "https://httpbin.org/bytes/1024"},
			{"Image", "https://httpbin.org/image/png"},
		}

		for _, test := range contentTypeTests {
			t.Run(test.name, func(t *testing.T) {
				requestBody := map[string]string{
					"url": test.url,
				}

				jsonBody, _ := json.Marshal(requestBody)
				
				client := &http.Client{
					Timeout: 30 * time.Second,
				}
				
				resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, test.url, response["url"])
				assert.Contains(t, response, "detected")
				assert.Contains(t, response, "content_type")

				t.Logf("%s analysis - Content-Type: %v", test.name, response["content_type"])
			})
		}
	})

	t.Run("Slow Response Handling", func(t *testing.T) {
		// Test with deliberately slow responses
		slowTests := []struct {
			name    string
			url     string
			timeout time.Duration
		}{
			{"1 Second Delay", "https://httpbin.org/delay/1", 30 * time.Second},
			{"5 Second Delay", "https://httpbin.org/delay/5", 30 * time.Second},
			{"10 Second Delay", "https://httpbin.org/delay/10", 30 * time.Second},
		}

		for _, test := range slowTests {
			t.Run(test.name, func(t *testing.T) {
				requestBody := map[string]string{
					"url": test.url,
				}

				jsonBody, _ := json.Marshal(requestBody)
				
				client := &http.Client{
					Timeout: test.timeout,
				}
				
				start := time.Now()
				resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
				duration := time.Since(start)
				
				require.NoError(t, err)
				defer resp.Body.Close()

				t.Logf("%s took %v", test.name, duration)

				// Should handle slow responses within timeout
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, test.url, response["url"])
			})
		}
	})

	t.Run("Memory Pressure During Analysis", func(t *testing.T) {
		// Test multiple concurrent requests to create memory pressure
		const numRequests = 10
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				requestBody := map[string]string{
					"url": "https://httpbin.org/bytes/1048576", // 1MB response
				}

				jsonBody, _ := json.Marshal(requestBody)
				
				client := &http.Client{
					Timeout: 60 * time.Second,
				}
				
				resp, err := client.Post(serverURL+"/v1/analyze", "application/json", bytes.NewBuffer(jsonBody))
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
			case <-time.After(120 * time.Second):
				t.Fatal("Timeout waiting for memory pressure test requests")
			}
		}

		// Should handle at least some requests successfully under memory pressure
		assert.GreaterOrEqual(t, successCount, numRequests/3, "Should handle at least 1/3 of requests under memory pressure")
		t.Logf("Successfully handled %d/%d requests under memory pressure", successCount, numRequests)
	})
}