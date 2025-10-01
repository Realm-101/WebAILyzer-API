# WebAIlyzer Lite API - Project Structure

This document describes the clean, minimal project structure of the WebAIlyzer Lite API.

## Directory Structure

```
.
├── .github/                     # GitHub workflows and templates
├── .kiro/                       # Kiro IDE configuration and specs
├── cmd/                         # Application entry points
│   └── webailyzer-api/         # Main API server
│       ├── main.go             # Application bootstrap and HTTP server
│       ├── main_test.go        # Main application tests
│       ├── memory_test.go      # Memory optimization tests
│       ├── resource_optimization_test.go  # Resource usage tests
│       └── timeout_test.go     # Timeout handling tests
├── examples/                    # Usage examples
│   ├── integration_examples.md # Integration examples documentation
│   └── main.go                 # Simple wappalyzer usage example
├── test/                        # Test suite
│   ├── integration/            # Integration tests
│   │   ├── docker_test.go      # Docker deployment tests
│   │   ├── edge_cases_test.go  # Edge case handling tests
│   │   ├── real_server_test.go # Real server integration tests
│   │   ├── webailyzer_lite_test.go # Core API tests
│   │   └── README.md           # Integration test documentation
│   ├── run_integration_tests.bat # Windows test runner
│   └── run_integration_tests.sh  # Linux/Mac test runner
├── .gitignore                   # Git ignore patterns
├── API_DOCUMENTATION.md         # Complete API reference
├── CHANGELOG.md                 # Version history and changes
├── CONTRIBUTING.md              # Contribution guidelines
├── deploy.bat                   # Windows deployment script
├── deploy.sh                    # Linux/Mac deployment script
├── DEPLOYMENT.md                # Comprehensive deployment guide
├── docker-compose.test.yml      # Docker Compose for testing
├── docker-compose.yml           # Docker Compose for development
├── Dockerfile                   # Docker image definition
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── LICENSE                      # MIT license
├── Makefile                     # Build automation
├── QUICK_START.md               # Quick deployment guide
├── README.md                    # Project overview and usage
├── test-docker.bat              # Windows Docker test script
└── test-docker.sh               # Linux/Mac Docker test script
```

## Key Components

### Core Application (`cmd/webailyzer-api/`)
- **main.go**: Complete HTTP server implementation with:
  - Health check endpoint (`/health`)
  - Website analysis endpoint (`/v1/analyze`)
  - Error handling and logging
  - Memory optimization
  - Request timeouts and resource management

### Testing (`test/`)
- **Integration tests**: End-to-end API testing
- **Docker tests**: Container deployment verification
- **Edge case tests**: Error handling and boundary conditions
- **Performance tests**: Memory and resource usage validation

### Documentation
- **README.md**: Project overview and quick start
- **DEPLOYMENT.md**: Comprehensive deployment guide
- **API_DOCUMENTATION.md**: Complete API reference
- **QUICK_START.md**: One-command deployment guide

### Deployment
- **Dockerfile**: Multi-stage build with security optimizations
- **docker-compose.yml**: Simple development deployment
- **deploy.sh/deploy.bat**: Automated deployment scripts
- **Makefile**: Build automation and common tasks

## Architecture Principles

### Simplicity
- Single binary with no external dependencies
- Minimal configuration required
- No database or persistent storage needed

### Performance
- Optimized HTTP client with connection pooling
- Memory-efficient request processing
- Garbage collection tuning for low resource usage

### Security
- Read-only container filesystem
- Non-root user execution
- Request size limits and timeouts
- Input validation and sanitization

### Reliability
- Comprehensive error handling
- Health check endpoints
- Graceful shutdown handling
- Resource cleanup and memory management

## Development Workflow

1. **Local Development**: Use `go run ./cmd/webailyzer-api` or `make run`
2. **Testing**: Run `make test` or use the integration test scripts
3. **Building**: Use `make build` or `go build ./cmd/webailyzer-api`
4. **Docker Testing**: Use `./test-docker.sh` or `test-docker.bat`
5. **Deployment**: Use `./deploy.sh` or `deploy.bat` for automated deployment

## Dependencies

### Runtime Dependencies
- **Go 1.24+**: Core runtime
- **github.com/projectdiscovery/wappalyzergo**: Technology detection
- **github.com/gorilla/mux**: HTTP routing
- **github.com/gorilla/handlers**: CORS and middleware
- **github.com/sirupsen/logrus**: Structured logging

### Development Dependencies
- **Docker**: For containerized deployment
- **Make**: For build automation (optional)
- **curl**: For API testing

## File Naming Conventions

- **Go files**: `snake_case.go`
- **Test files**: `*_test.go`
- **Documentation**: `UPPERCASE.md`
- **Scripts**: `kebab-case.sh` or `kebab-case.bat`
- **Configuration**: `kebab-case.yml` or `kebab-case.yaml`

This structure prioritizes simplicity, maintainability, and ease of deployment while providing comprehensive testing and documentation.