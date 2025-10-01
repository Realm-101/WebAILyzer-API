# WebAIlyzer Lite API

A minimal, lightweight web technology detection API built with Go. Provides simple HTTP endpoints for analyzing websites and detecting technologies using the wappalyzer engine.

## Features

- **Technology Detection**: Identify web technologies, frameworks, and libraries used by websites
- **Simple HTTP API**: Two endpoints - health check and website analysis
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **Lightweight**: Minimal dependencies and resource usage
- **Fast Response**: Quick analysis with appropriate timeouts

## Quick Start

### One-Command Deployment (Recommended)

**Linux/Mac:**
```bash
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api && ./deploy.sh
```

**Windows:**
```cmd
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api && deploy.bat
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api

# Start with Docker Compose
docker-compose up -d

# The API will be available at http://localhost:8080
curl http://localhost:8080/health
```

📋 **See [QUICK_START.md](QUICK_START.md) for the fastest deployment method**

### Building from Source

```bash
# Install dependencies
go mod download

# Build the application
go build -o webailyzer-api ./cmd/webailyzer-api

# Run the application
./webailyzer-api
```

## API Usage

The API provides two simple endpoints with no authentication required:

### Health Check

Check if the API is running:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok"
}
```

### Website Analysis

Analyze a website to detect technologies:

```bash
curl -X POST http://localhost:8080/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com"
  }'
```

Response:
```json
{
  "url": "https://example.com",
  "detected": {
    "Nginx": {
      "categories": ["Web servers"],
      "confidence": 100,
      "version": "",
      "icon": "Nginx.svg",
      "website": "http://nginx.org/en",
      "cpe": "cpe:/a:nginx:nginx"
    },
    "Bootstrap": {
      "categories": ["UI frameworks"],
      "confidence": 100,
      "version": "4.3.1",
      "icon": "Bootstrap.svg",
      "website": "https://getbootstrap.com"
    }
  },
  "content_type": "text/html; charset=utf-8"
}
```

## Configuration

No configuration is required. The API runs on port 8080 by default.

### Docker Compose Configuration

The included `docker-compose.yml` provides a simple setup:

```yaml
version: '3.8'
services:
  webailyzer-api:
    build: .
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## API Endpoints

The API provides two simple endpoints:

- `GET /health` - Health check endpoint
- `POST /v1/analyze` - Analyze a website for technology detection

## Development

### Running Tests

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Building

```bash
# Build for current platform
go build -o webailyzer-api ./cmd/webailyzer-api

# Build for Linux (for Docker)
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webailyzer-api ./cmd/webailyzer-api
```

## Deployment

For detailed deployment instructions, environment configuration, and troubleshooting, see [DEPLOYMENT.md](DEPLOYMENT.md).

### Quick Docker Deployment

```bash
# Build Docker image
docker build -t webailyzer-lite-api .

# Run container
docker run -p 8080:8080 webailyzer-lite-api
```

### Health Checks

The API includes a health check endpoint at `/health` that returns:

```json
{
  "status": "ok"
}
```

This endpoint is used by Docker health checks and load balancers to verify the service is running.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
