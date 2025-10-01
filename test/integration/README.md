# Integration Tests for WebAIlyzer Lite API

This directory contains comprehensive integration tests for the WebAIlyzer Lite API. The tests are designed to verify the complete functionality of the API endpoints, error handling, and Docker container behavior.

## Test Structure

### Test Files

1. **`webailyzer_lite_test.go`** - Basic integration tests using a mock server
   - Health endpoint testing
   - Analyze endpoint testing with various scenarios
   - Error handling validation
   - Concurrent request testing
   - Response header validation

2. **`real_server_test.go`** - Tests against the actual compiled server
   - Real HTTP requests to live endpoints
   - Actual technology detection using Wappalyzer
   - Network error handling
   - Large response handling
   - SSL certificate validation
   - CORS functionality

3. **`docker_test.go`** - Docker container integration tests
   - Docker image building and container startup
   - Health checks within containers
   - Resource limit testing
   - Container log validation
   - Multi-container scenarios

4. **`edge_cases_test.go`** - Edge cases and unusual scenarios
   - Very long URLs
   - Special characters in URLs
   - International domain names
   - Different HTTP methods
   - Various content types
   - Malformed JSON payloads
   - Different URL schemes
   - Redirect handling
   - Slow response handling
   - Memory pressure testing

## Test Categories

### Basic Functionality Tests
- Health endpoint (`/health`)
- Analysis endpoint (`/v1/analyze`)
- Request/response validation
- Error response structure

### Real-World Scenario Tests
- Analysis of actual websites (httpbin.org, example.com)
- Network timeout handling
- Invalid SSL certificate handling
- Large response processing
- User agent and header validation

### Error Handling Tests
- Invalid URL formats
- Unreachable domains
- HTTP error status codes (404, 500, etc.)
- Malformed JSON requests
- Missing required fields
- Unsupported HTTP methods

### Performance and Reliability Tests
- Concurrent request handling
- Memory pressure scenarios
- Slow response handling
- Resource limit validation
- Container resource constraints

### Docker Integration Tests
- Container build and startup
- Health check functionality
- Network connectivity within containers
- Resource usage monitoring
- Container log analysis

## Running the Tests

### Prerequisites

1. **Go 1.21+** - Required for running the tests
2. **Docker** - Required for Docker integration tests (optional)
3. **Internet connection** - Required for real-world URL testing

### Quick Start

```bash
# Run basic integration tests (no Docker required)
./test/run_integration_tests.sh basic

# Run all integration tests including Docker
./test/run_integration_tests.sh integration

# Run only Docker tests
./test/run_integration_tests.sh docker
```

### Windows

```cmd
# Run basic integration tests
test\run_integration_tests.bat basic

# Run all integration tests
test\run_integration_tests.bat integration

# Run only Docker tests
test\run_integration_tests.bat docker
```

### Manual Test Execution

```bash
# Run all integration tests
go test -v ./test/integration/... -timeout=30m

# Run specific test file
go test -v ./test/integration/ -run TestHealthEndpoint

# Run tests excluding Docker (short mode)
go test -v ./test/integration/... -short

# Run with race detection
go test -v -race ./test/integration/...
```

## Test Configuration

### Environment Variables

The tests automatically detect the environment and adjust behavior:

- **Short mode** (`-short` flag): Skips Docker tests and long-running tests
- **Timeout**: Default 30 minutes for full test suite
- **Ports**: Tests use ports 18080, 18081 to avoid conflicts

### Test Data

Tests use the following external services for real-world testing:
- `httpbin.org` - HTTP testing service
- `example.com` - Simple HTML page
- Various test URLs for edge cases

## Test Output

### Success Indicators
- All endpoints return expected status codes
- Response structures match API specification
- Error responses include proper error types and messages
- Docker containers start and respond correctly
- Memory usage stays within acceptable limits

### Failure Scenarios
Tests are designed to handle and validate:
- Network connectivity issues
- Service unavailability
- Invalid input data
- Resource constraints
- Container failures

## Continuous Integration

The integration tests are designed to run in CI/CD environments:

```yaml
# Example GitHub Actions step
- name: Run Integration Tests
  run: |
    ./test/run_integration_tests.sh basic
    
# With Docker support
- name: Run Full Integration Tests
  run: |
    ./test/run_integration_tests.sh integration
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Tests use ports 18080, 18081. Ensure these are available.
2. **Docker not available**: Use `basic` command to skip Docker tests.
3. **Network timeouts**: Some tests require internet connectivity.
4. **Build failures**: Ensure Go modules are properly downloaded (`go mod tidy`).

### Debug Mode

Enable verbose logging:
```bash
go test -v ./test/integration/... -args -test.v
```

### Test Isolation

Each test is designed to be independent:
- Tests use different ports to avoid conflicts
- Docker containers are automatically cleaned up
- No persistent state between tests
- Temporary files are cleaned up automatically

## Coverage

Integration tests provide coverage for:
- ✅ HTTP endpoint functionality
- ✅ Request/response handling
- ✅ Error scenarios and edge cases
- ✅ Docker container behavior
- ✅ Real-world usage patterns
- ✅ Performance under load
- ✅ Resource management
- ✅ Network error handling

## Contributing

When adding new integration tests:

1. Follow the existing test structure and naming conventions
2. Include both positive and negative test cases
3. Add appropriate timeouts for network operations
4. Clean up resources (servers, containers) in defer statements
5. Use descriptive test names and comments
6. Test both success and failure scenarios
7. Validate response structure and content
8. Consider edge cases and unusual inputs

## Performance Benchmarks

The integration tests also serve as performance benchmarks:
- Response time measurements
- Memory usage tracking
- Concurrent request handling
- Resource utilization monitoring

Results are logged during test execution for performance regression detection.