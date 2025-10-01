# Requirements Document

## Introduction

WebAIlyzer Lite API is a minimal, cost-effective web technology detection service that fixes the current build issues and provides a simple HTTP API for analyzing websites. The system focuses on getting a working API running quickly with minimal dependencies and infrastructure costs, while maintaining the core wappalyzer technology detection functionality.

## Requirements

### Requirement 1: Fix Build and Module Issues

**User Story:** As a developer, I want the project to build and run successfully, so that I can deploy and use the API without configuration issues.

#### Acceptance Criteria

1. WHEN running `go build` THEN the system SHALL compile successfully without module path errors
2. WHEN the go.mod file is updated THEN it SHALL point to the correct repository path instead of upstream wappalyzergo
3. WHEN building the Docker image THEN it SHALL use the correct entrypoint path that exists in the repository
4. WHEN running the application THEN it SHALL start without missing dependency errors
5. IF there are missing files referenced in build scripts THEN they SHALL be either created or removed from the build process

### Requirement 2: Simple HTTP API Server

**User Story:** As a developer integrating the API, I want a basic HTTP server with essential endpoints, so that I can analyze websites programmatically.

#### Acceptance Criteria

1. WHEN a POST request is made to /v1/analyze THEN the system SHALL return wappalyzer technology detection results for the provided URL
2. WHEN a GET request is made to /health THEN the system SHALL return a simple health status response
3. WHEN the server starts THEN it SHALL listen on port 8080 and log startup information
4. WHEN invalid JSON is sent to /v1/analyze THEN the system SHALL return a 400 error with clear error message
5. IF a URL cannot be fetched THEN the system SHALL return a 502 error with appropriate error details

### Requirement 3: Minimal Technology Detection

**User Story:** As a user, I want to detect technologies used by websites, so that I can understand the tech stack of any given URL.

#### Acceptance Criteria

1. WHEN analyzing a URL THEN the system SHALL fetch the webpage content and HTTP headers
2. WHEN processing webpage data THEN the system SHALL use wappalyzer fingerprinting to detect technologies
3. WHEN returning results THEN the system SHALL include the detected technologies in a structured JSON format
4. WHEN analysis completes THEN the system SHALL include the original URL and content type in the response
5. IF the webpage cannot be accessed THEN the system SHALL return an error indicating the fetch failure

### Requirement 4: Basic Error Handling

**User Story:** As a developer integrating the API, I want clear error responses, so that I can handle failures appropriately in my application.

#### Acceptance Criteria

1. WHEN invalid requests are received THEN the system SHALL return HTTP 400 with JSON error details
2. WHEN URLs cannot be fetched THEN the system SHALL return HTTP 502 with gateway error information
3. WHEN internal errors occur THEN the system SHALL return HTTP 500 with generic error message
4. WHEN request timeouts occur THEN the system SHALL return appropriate timeout error responses
5. IF the wappalyzer engine fails to initialize THEN the system SHALL return HTTP 500 with initialization error

### Requirement 5: Docker Deployment Support

**User Story:** As a system administrator, I want to deploy the API using Docker, so that I can run it consistently across different environments.

#### Acceptance Criteria

1. WHEN building the Docker image THEN it SHALL compile the Go application successfully
2. WHEN running the Docker container THEN it SHALL expose port 8080 for HTTP traffic
3. WHEN the container starts THEN the health check SHALL verify the /health endpoint responds correctly
4. WHEN the Dockerfile is built THEN it SHALL use multi-stage builds to minimize image size
5. IF the health check fails THEN Docker SHALL report the container as unhealthy

### Requirement 6: Configuration Consistency

**User Story:** As a developer, I want consistent configuration between documentation and implementation, so that the API behaves as documented.

#### Acceptance Criteria

1. WHEN the README documents endpoints THEN they SHALL match the actual implemented endpoints
2. WHEN the Dockerfile references health check paths THEN they SHALL match the implemented health endpoint
3. WHEN configuration files are referenced THEN they SHALL either exist in the repository or be removed from references
4. WHEN the quick start guide is followed THEN it SHALL result in a working deployment
5. IF there are discrepancies between docs and code THEN they SHALL be resolved to match the simpler implementation

### Requirement 7: HTTP Client Safety

**User Story:** As a system administrator, I want the API to handle external HTTP requests safely, so that it doesn't become a vector for abuse or hang indefinitely.

#### Acceptance Criteria

1. WHEN making HTTP requests to analyze URLs THEN the system SHALL use reasonable timeouts (15 seconds max)
2. WHEN connecting to external hosts THEN the system SHALL limit connection time to prevent hanging
3. WHEN processing responses THEN the system SHALL limit response body size to prevent memory exhaustion
4. WHEN handling redirects THEN the system SHALL follow a reasonable number of redirects (max 10)
5. IF external requests take too long THEN the system SHALL timeout and return an appropriate error

### Requirement 8: Minimal Resource Usage

**User Story:** As a cost-conscious user, I want the API to use minimal resources, so that I can run it cheaply on basic hosting.

#### Acceptance Criteria

1. WHEN the application runs THEN it SHALL use minimal memory footprint suitable for shared hosting
2. WHEN processing requests THEN it SHALL not require persistent storage or databases
3. WHEN handling concurrent requests THEN it SHALL manage resources efficiently without memory leaks
4. WHEN idle THEN the system SHALL consume minimal CPU and memory resources
5. IF resource usage grows THEN it SHALL be due to legitimate request processing, not memory leaks