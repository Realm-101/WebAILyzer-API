# WebAIlyzer Lite API

A comprehensive website analysis API that provides performance metrics, SEO analysis, accessibility checks, security assessments, and AI-powered insights. Built with Go for high performance and scalability.

## Features

- **Comprehensive Analysis**: Performance, SEO, accessibility, security, and technology detection
- **AI-Powered Insights**: Automated recommendations and optimization suggestions
- **Batch Processing**: Analyze multiple URLs simultaneously
- **Real-time Metrics**: Track and analyze website performance over time
- **Export Capabilities**: Generate reports in multiple formats (PDF, CSV, JSON)
- **Event Tracking**: Monitor user interactions and behavior
- **Rate Limiting**: Built-in API rate limiting and authentication
- **Caching**: Redis-based caching for improved performance
- **Monitoring**: Prometheus metrics and health checks
- **Scalable Architecture**: Microservices-ready with Docker support

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/projectdiscovery/wappalyzergo.git
cd wappalyzergo

# Start with Docker Compose
docker-compose up -d

# The API will be available at http://localhost:8080
curl http://localhost:8080/health
```

### Building from Source

```bash
# Install dependencies
go mod download

# Build the application
make build

# Run the application
./bin/webailyzer-api
```

## API Usage

### Authentication

All API requests require authentication using a Bearer token:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:8080/api/v1/analyze
```

### Basic Analysis

Analyze a single website:

```bash
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "options": {
      "include_performance": true,
      "include_seo": true,
      "include_accessibility": true,
      "include_security": true
    }
  }'
```

### Batch Analysis

Analyze multiple URLs at once:

```bash
curl -X POST http://localhost:8080/api/v1/batch \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://example1.com", "https://example2.com"],
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "options": {
      "include_performance": true,
      "include_seo": true
    }
  }'
```

### Generate Insights

Get AI-powered optimization recommendations:

```bash
curl -X POST http://localhost:8080/api/v1/insights/generate \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

## Configuration

### Environment Variables

```bash
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=webailyzer
DATABASE_PASSWORD=your_password
DATABASE_NAME=webailyzer

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# Application Configuration
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=production

# Rate Limiting
RATE_LIMIT_DEFAULT=1000
RATE_LIMIT_WINDOW_DURATION=1h
```

### Docker Compose Configuration

The included `docker-compose.yml` provides a complete setup with PostgreSQL and Redis:

```yaml
version: '3.8'
services:
  webailyzer-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
```

## API Endpoints

### Core Analysis
- `POST /api/v1/analyze` - Analyze a single URL
- `POST /api/v1/batch` - Batch analyze multiple URLs
- `GET /api/v1/analysis` - Get analysis history

### Insights & Recommendations
- `POST /api/v1/insights/generate` - Generate AI insights
- `GET /api/v1/insights` - Get insights for workspace
- `PUT /api/v1/insights/{id}/status` - Update insight status

### Metrics & Analytics
- `GET /api/v1/metrics` - Get aggregated metrics
- `GET /api/v1/events` - Get tracked events
- `POST /api/v1/events/track` - Track custom events

### Data Export
- `POST /api/v1/export` - Export analysis data
- `GET /api/v1/export/{id}` - Download export file

### System
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

## Development

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# Benchmark tests
make test-benchmarks

# All tests
make test-all
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create
```

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Generate documentation
make docs
```

## Monitoring & Observability

### Prometheus Metrics

The API exposes Prometheus metrics at `/metrics`:

- Request duration and count
- Database connection pool stats
- Cache hit/miss rates
- Analysis processing times
- Error rates by endpoint

### Health Checks

Multiple health check endpoints:

- `/health` - Basic health status
- `/health/detailed` - Detailed component health
- `/health/db` - Database connectivity
- `/health/redis` - Redis connectivity

### Logging

Structured logging with configurable levels:

```bash
# Set log level
export LOG_LEVEL=debug

# JSON format for production
export LOG_FORMAT=json
```

## Deployment

### Production Deployment

See [DEPLOYMENT.md](API_DOCUMENTATION.md#deployment-and-configuration-guide) for detailed deployment instructions including:

- Docker and Kubernetes configurations
- Database setup and optimization
- Load balancing and high availability
- Security configuration
- Backup and recovery procedures

### Scaling Considerations

- **Horizontal Scaling**: Stateless design allows easy scaling
- **Database**: Use read replicas and connection pooling
- **Caching**: Redis cluster for distributed caching
- **Load Balancing**: Nginx or cloud load balancers

## Documentation

- [API Documentation](API_DOCUMENTATION.md) - Complete API reference
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [Project Structure](PROJECT_STRUCTURE.md) - Codebase organization

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
