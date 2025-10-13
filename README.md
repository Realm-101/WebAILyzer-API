# WebAIlyzer Lite API

A minimal, lightweight web technology detection API built with Go. Provides simple HTTP endpoints for analyzing websites and detecting technologies using the wappalyzer engine.

> 🧹 **Clean & Simple**: This repository has been cleaned and optimized for minimal complexity and maximum usability.

## Features

- **Technology Detection**: Identify web technologies, frameworks, and libraries used by websites
- **Simple HTTP API**: Two endpoints - health check and website analysis
- **API Key Authentication**: Secure access with mandatory API keys to prevent abuse
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **Lightweight**: Minimal dependencies and resource usage (runs in <256MB RAM)
- **Fast Response**: Quick analysis with appropriate timeouts
- **Production Ready**: Includes health checks, logging, and error handling

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

# Run the application (set API_KEYS for production)
export API_KEYS="your-secret-api-key"
./webailyzer-api
```

## API Usage

Access to the analysis endpoint is protected by API keys. You must provide a valid key in the `Authorization` header. The health check endpoint does not require authentication.

### Health Check

Check if the API is running (no authentication required):

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

Analyze a website to detect technologies. Requires a valid API key.

```bash
curl -X POST http://localhost:8080/v1/analyze \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
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

The API is configured via environment variables.

### Environment Variables
- `PORT`: The port the server listens on. Defaults to `8080`.
- `API_KEYS`: A comma-separated list of valid API keys for authentication. **Required for production.**

Example:
```
PORT=8080
API_KEYS=key1_secret,key2_secret
```

### Docker Compose Configuration

Update the `docker-compose.yml` to include your API keys using an environment file or directly.

```yaml
version: '3.8'
services:
  webailyzer-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - API_KEYS=your-secret-key-here
    restart: unless-stopped
```

## API Endpoints

- `GET /health` - Health check endpoint (unauthenticated)
- `POST /v1/analyze` - Analyze a website for technology detection (requires authentication)

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

# Run container with API key
docker run -p 8080:8080 -e API_KEYS="your-secret-key" webailyzer-lite-api
```

### Health Checks

The API includes a health check endpoint at `/health` that returns:

```json
{
  "status": "ok"
}
```

This endpoint is used by Docker health checks and load balancers to verify the service is running.

## Project Structure

The project follows a clean, minimal structure focused on simplicity and maintainability:

```
├── cmd/webailyzer-api/     # Main application
├── test/                   # Integration tests
├── examples/               # Usage examples
├── deploy.sh/deploy.bat    # Deployment scripts
├── DEPLOYMENT.md           # Deployment guide
├── API_DOCUMENTATION.md    # API reference
└── QUICK_START.md          # Quick start guide
```

📋 **See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for detailed project organization**

## Documentation

- **[QUICK_START.md](QUICK_START.md)** - One-command deployment
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Comprehensive deployment guide  
- **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** - Complete API reference
- **[PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)** - Project organization
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines
- **[CHANGELOG.md](CHANGELOG.md)** - Version history

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `make test`
6. Test Docker deployment: `./test-docker.sh`
7. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.