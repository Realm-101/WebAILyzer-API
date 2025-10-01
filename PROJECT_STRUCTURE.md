# WebAIlyzer Lite API - Project Structure

This document describes the project structure and organization of the WebAIlyzer Lite API.

## Directory Structure

```
.
├── cmd/                          # Application entry points
│   └── webailyzer-api/          # Main API server
│       └── main.go              # Application bootstrap
├── internal/                     # Private application code
│   ├── cache/                   # Caching layer (Redis)
│   │   ├── redis.go             # Redis client implementation
│   │   └── service.go           # Cache service interface
│   ├── config/                  # Configuration management
│   │   └── config.go            # Configuration structures and loading
│   ├── database/                # Database management
│   │   ├── connection.go        # Database connection handling
│   │   ├── migrations.go        # Migration management
│   │   ├── maintenance.go       # Database maintenance tasks
│   │   ├── performance.go       # Performance optimization
│   │   └── migrations/          # SQL migration files
│   ├── errors/                  # Error handling
│   │   └── errors.go            # Custom error types and handling
│   ├── handlers/                # HTTP request handlers
│   │   ├── analysis.go          # Analysis endpoints
│   │   ├── event.go             # Event tracking endpoints
│   │   ├── export.go            # Data export endpoints
│   │   ├── health.go            # Health check endpoints
│   │   ├── insights.go          # Insights endpoints
│   │   └── metrics.go           # Metrics endpoints
│   ├── logging/                 # Logging infrastructure
│   │   └── logger.go            # Structured logging setup
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go              # Authentication middleware
│   │   ├── error.go             # Error handling middleware
│   │   ├── metrics.go           # Metrics collection middleware
│   │   └── ratelimit.go         # Rate limiting middleware
│   ├── models/                  # Data models and structures
│   │   ├── analysis.go          # Analysis-related models
│   │   ├── insight.go           # Insight models
│   │   ├── metrics.go           # Metrics models
│   │   ├── session.go           # Session models
│   │   └── workspace.go         # Workspace models
│   ├── monitoring/              # Monitoring and observability
│   │   ├── metrics.go           # Prometheus metrics
│   │   └── service.go           # Monitoring service
│   ├── repositories/            # Data access layer
│   │   ├── interfaces.go        # Repository interfaces
│   │   └── postgres/            # PostgreSQL implementations
│   │       ├── analysis.go      # Analysis repository
│   │       ├── event.go         # Event repository
│   │       ├── insight.go       # Insight repository
│   │       ├── metrics.go       # Metrics repository
│   │       ├── session.go       # Session repository
│   │       └── workspace.go     # Workspace repository
│   └── services/                # Business logic layer
│       ├── analysis.go          # Analysis service implementation
│       ├── event.go             # Event tracking service
│       ├── export.go            # Export service
│       ├── insights.go          # Insights generation service
│       ├── insights_job.go      # Background insights processing
│       ├── interfaces.go        # Service interfaces
│       ├── metrics.go           # Metrics aggregation service
│       └── analyzers/           # Analysis engines
│           ├── accessibility.go # Accessibility analyzer
│           ├── performance.go   # Performance analyzer
│           ├── security.go      # Security analyzer
│           ├── seo.go           # SEO analyzer
│           └── technology.go    # Technology detection
├── test/                        # Test files
│   ├── benchmarks/              # Performance benchmarks
│   │   └── performance_test.go  # API performance tests
│   ├── integration/             # Integration tests
│   │   ├── e2e_test.go          # End-to-end tests
│   │   └── error_scenarios_test.go # Error handling tests
│   └── run_integration_tests.sh # Test runner script
├── monitoring/                  # Monitoring configuration
│   ├── grafana-dashboard.json   # Grafana dashboard
│   ├── prometheus.yml           # Prometheus configuration
│   └── webailyzer_rules.yml     # Alerting rules
├── API_DOCUMENTATION.md         # Complete API documentation
├── TROUBLESHOOTING.md           # Troubleshooting guide
├── docker-compose.yml           # Development environment
├── docker-compose.test.yml      # Test environment
├── Dockerfile                   # Container build instructions
├── Makefile                     # Build and development tasks
├── go.mod                       # Go module definition
└── README.md                    # Project overview and setup
```

## Architecture Layers

### 1. Handlers Layer (`internal/handlers/`)
- HTTP request/response handling
- Request validation and parsing
- Response formatting
- Route registration

### 2. Services Layer (`internal/services/`)
- Business logic implementation
- Orchestration of multiple repositories
- Data transformation and validation
- External service integration

### 3. Repository Layer (`internal/repositories/`)
- Data access abstraction
- Database operations (CRUD)
- Query optimization
- Transaction management

### 4. Models Layer (`internal/models/`)
- Data structures and types
- Validation rules
- JSON serialization tags
- Database mapping

### 5. Infrastructure Layer
- **Config** (`internal/config/`): Configuration management
- **Database** (`internal/database/`): Database connection handling
- **Cache** (`internal/cache/`): Caching layer implementation

## Configuration

The application uses a hierarchical configuration system:

1. **Default values** (set in code)
2. **Configuration file** (`config.yaml`)
3. **Environment variables** (prefixed with `WEBAILYZER_`)

Environment variables take precedence over config file values.

## Development Workflow

### Prerequisites
- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Local Development
```bash
# Install dependencies
make deps

# Format and lint code
make fmt lint

# Run tests
make test

# Start local development server
make run
```

### Docker Development
```bash
# Start all services (PostgreSQL, Redis, API)
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Database Migrations

Database schema changes are managed through migration files in the `migrations/` directory.

```bash
# Create a new migration
make migrate-create

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Testing Strategy

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database operations and external services
- **API Tests**: Test HTTP endpoints end-to-end

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Deployment

### Docker Deployment
```bash
# Build container
make docker-build

# Run container
make docker-run
```

### Configuration for Production
Set the following environment variables:

- `WEBAILYZER_DATABASE_HOST`
- `WEBAILYZER_DATABASE_PASSWORD`
- `WEBAILYZER_REDIS_HOST`
- `WEBAILYZER_AUTH_JWT_SECRET`
- `WEBAILYZER_LOGGING_LEVEL`

## Implementation Status

The WebAIlyzer Lite API is fully implemented with the following features:

### ✅ Completed Features

1. **Core Analysis Engine**
   - Technology detection using Wappalyzer fingerprints
   - Performance metrics collection (Core Web Vitals)
   - SEO analysis (meta tags, headings, links)
   - Accessibility checks (WCAG compliance)
   - Security assessment (SSL, headers, vulnerabilities)

2. **AI-Powered Insights**
   - Automated insight generation from analysis data
   - Performance optimization recommendations
   - SEO improvement suggestions
   - Accessibility enhancement tips
   - Security vulnerability alerts

3. **Data Management**
   - PostgreSQL database with optimized schema
   - Redis caching for improved performance
   - Database migrations and maintenance
   - Data export in multiple formats (PDF, CSV, JSON)

4. **API Features**
   - RESTful API with comprehensive endpoints
   - Authentication and workspace management
   - Rate limiting and request validation
   - Batch processing capabilities
   - Event tracking and analytics

5. **Monitoring & Observability**
   - Prometheus metrics collection
   - Structured logging with configurable levels
   - Health checks and system monitoring
   - Performance profiling endpoints
   - Grafana dashboard configuration

6. **Testing & Quality**
   - Comprehensive unit test suite
   - Integration tests with database
   - End-to-end API testing
   - Performance benchmarks
   - Error scenario testing

7. **Deployment & Operations**
   - Docker containerization
   - Docker Compose for development
   - Kubernetes deployment manifests
   - Production-ready configuration
   - Backup and recovery procedures

### 🏗️ Architecture Highlights

- **Clean Architecture**: Separation of concerns with clear layer boundaries
- **Dependency Injection**: Testable and maintainable code structure
- **Interface-Driven Design**: Easy to mock and test components
- **Microservices Ready**: Stateless design for horizontal scaling
- **Performance Optimized**: Caching, connection pooling, and efficient queries
- **Security First**: Authentication, rate limiting, and input validation

### 📊 Key Metrics

- **Test Coverage**: Comprehensive test suite with unit and integration tests
- **Performance**: Optimized for high throughput and low latency
- **Scalability**: Horizontal scaling support with load balancing
- **Reliability**: Error handling, circuit breakers, and graceful degradation
- **Observability**: Metrics, logging, and tracing for production monitoring

The project follows Go best practices and clean architecture principles, making it maintainable, testable, and production-ready.