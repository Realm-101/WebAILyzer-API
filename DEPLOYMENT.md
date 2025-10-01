# WebAIlyzer Lite API - Deployment Guide

This guide provides comprehensive instructions for deploying and running the WebAIlyzer Lite API in various environments.

## Table of Contents

- [Quick Start with Docker](#quick-start-with-docker)
- [Local Development Setup](#local-development-setup)
- [Production Deployment](#production-deployment)
- [Environment Variables](#environment-variables)
- [Configuration Options](#configuration-options)
- [Health Checks](#health-checks)
- [Troubleshooting](#troubleshooting)

## Quick Start with Docker

The fastest way to get WebAIlyzer Lite API running is using the provided deployment scripts:

### Prerequisites

- Docker 20.10+ installed
- Docker Compose 2.0+ (optional, for easier management)

### Using Deployment Scripts (Recommended)

The repository includes automated deployment scripts that handle building, running, and testing the API:

#### Linux/Mac (`deploy.sh`)
```bash
# Clone the repository
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api

# Run the deployment script
./deploy.sh

# Or with custom options
./deploy.sh -p 9000 -m 512m -c 1.0

# View all options
./deploy.sh --help
```

#### Windows (`deploy.bat`)
```cmd
# Clone the repository
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api

# Run the deployment script
deploy.bat

# Or with custom options
deploy.bat -p 9000 -m 512m -c 1.0

# View all options
deploy.bat --help
```

#### Deployment Script Features

The deployment scripts automatically:
- Check Docker availability
- Build the Docker image
- Stop and remove any existing container
- Start a new container with optimized settings
- Wait for the service to be ready
- Test both API endpoints
- Display deployment information and management commands

#### Script Options

| Option | Description | Default |
|--------|-------------|---------|
| `-p, --port` | Port to expose | 8080 |
| `-m, --memory` | Memory limit | 256m |
| `-c, --cpu` | CPU limit | 0.5 |
| `-n, --name` | Container name | webailyzer-api |
| `-i, --image` | Image name | webailyzer-lite-api |
| `--no-test` | Skip API testing | false |
| `--cleanup-only` | Only cleanup existing container | false |

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# Verify it's running
curl http://localhost:8080/health
```

### Using Docker directly

```bash
# Build the image
docker build -t webailyzer-lite-api .

# Run the container
docker run -d \
  --name webailyzer-api \
  -p 8080:8080 \
  --restart unless-stopped \
  webailyzer-lite-api

# Check logs
docker logs webailyzer-api
```

## Local Development Setup

### Prerequisites

- Go 1.24 or later
- Git

### Setup Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/webailyzer-lite-api.git
   cd webailyzer-lite-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o webailyzer-api ./cmd/webailyzer-api
   ```

4. **Run the application**
   ```bash
   ./webailyzer-api
   ```

5. **Verify it's working**
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Test analysis
   curl -X POST http://localhost:8080/v1/analyze \
     -H "Content-Type: application/json" \
     -d '{"url":"https://example.com"}'
   ```

### Development Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
./test/run_integration_tests.sh  # Linux/Mac
./test/run_integration_tests.bat # Windows

# Format code
go fmt ./...

# Lint code (if golangci-lint is installed)
golangci-lint run
```

## Production Deployment

### Docker Production Deployment

For production environments, consider these additional configurations:

#### 1. Using Docker with Resource Limits

```bash
docker run -d \
  --name webailyzer-api \
  -p 8080:8080 \
  --restart unless-stopped \
  --memory="256m" \
  --cpus="0.5" \
  --read-only \
  --tmpfs /tmp \
  webailyzer-lite-api
```

#### 2. Docker Compose for Production

Create a `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  api:
    build: .
    container_name: webailyzer-lite-api
    ports:
      - "8080:8080"
    restart: unless-stopped
    read_only: true
    tmpfs:
      - /tmp
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
        reservations:
          memory: 128M
          cpus: '0.25'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      start_period: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

Run with:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Reverse Proxy Setup

#### Nginx Configuration

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }
}
```

#### Traefik Configuration

```yaml
# docker-compose.yml with Traefik
version: '3.8'

services:
  api:
    build: .
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webailyzer.rule=Host(`your-domain.com`)"
      - "traefik.http.services.webailyzer.loadbalancer.server.port=8080"
      - "traefik.http.routers.webailyzer.middlewares=ratelimit"
      - "traefik.http.middlewares.ratelimit.ratelimit.burst=10"
```

## Environment Variables

The WebAIlyzer Lite API currently uses hardcoded configuration values for simplicity and minimal setup requirements. The application runs on port 8080 with predefined timeouts and logging settings.

### Current Configuration (Hardcoded)

| Setting | Value | Description |
|---------|-------|-------------|
| HTTP Port | `8080` | Server listening port |
| Log Level | `info` | Logging level (JSON format) |
| HTTP Client Timeout | `15s` | Timeout for fetching external URLs |
| Connection Timeout | `5s` | Connection timeout for HTTP requests |
| Max Response Size | `5MB` | Maximum response body size to process |
| Read Timeout | `10s` | HTTP server read timeout |
| Write Timeout | `30s` | HTTP server write timeout |
| Idle Timeout | `60s` | HTTP server idle timeout |

### Future Environment Variable Support

If you need to customize these values, you can modify the source code in `cmd/webailyzer-api/main.go` or submit a feature request for environment variable support. The following variables could be implemented:

```bash
# Potential future environment variables
export PORT=8080
export LOG_LEVEL=info
export HTTP_TIMEOUT=15s
export CONNECTION_TIMEOUT=5s
export MAX_RESPONSE_SIZE=5MB
```

## Configuration Options

### HTTP Server Configuration

The server uses these default timeouts (hardcoded for simplicity):

- **Read Timeout**: 10 seconds
- **Write Timeout**: 30 seconds  
- **Idle Timeout**: 60 seconds

### HTTP Client Configuration

For fetching external URLs:

- **Request Timeout**: 15 seconds
- **Connection Timeout**: 5 seconds
- **TLS Handshake Timeout**: 5 seconds
- **Max Idle Connections**: 10
- **Max Redirects**: 10

### Resource Limits

Recommended resource limits:

- **Memory**: 256MB (can run on 128MB)
- **CPU**: 0.5 cores
- **Disk**: Minimal (stateless application)

## Health Checks

The API provides a health check endpoint for monitoring:

### Endpoint
```
GET /health
```

### Response
```json
{
  "status": "ok"
}
```

### Docker Health Check

The Dockerfile includes a built-in health check:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1
```

### Kubernetes Health Check

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: webailyzer-api
    image: webailyzer-lite-api
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 30
    readinessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
```

## Troubleshooting

### Common Issues

#### 1. Application Won't Start

**Symptoms**: Container exits immediately or binary fails to run

**Solutions**:
```bash
# Check logs
docker logs webailyzer-api

# Common fixes:
# - Ensure port 8080 is not already in use
sudo netstat -tlnp | grep :8080

# - Check if binary has execute permissions
chmod +x webailyzer-api

# - Verify Go version (requires 1.24+)
go version
```

#### 2. Build Failures

**Symptoms**: `go build` or `docker build` fails

**Solutions**:
```bash
# Clean module cache
go clean -modcache
go mod download

# Verify module path in go.mod
cat go.mod

# For Docker builds, check Dockerfile path
docker build -t webailyzer-lite-api .
```

#### 3. Health Check Failures

**Symptoms**: Docker reports container as unhealthy

**Solutions**:
```bash
# Test health endpoint manually
curl -v http://localhost:8080/health

# Check if curl is available in container
docker exec webailyzer-api curl --version

# Verify container is listening on correct port
docker exec webailyzer-api netstat -tlnp
```

#### 4. Analysis Endpoint Errors

**Symptoms**: `/v1/analyze` returns errors

**Common Error Responses**:

```json
// 400 Bad Request - Invalid JSON or missing URL
{
  "error": "invalid request: url required"
}

// 502 Bad Gateway - Cannot fetch URL
{
  "error": "fetch failed"
}

// 500 Internal Server Error - Wappalyzer initialization failed
{
  "error": "wappalyzer init failed"
}
```

**Solutions**:
```bash
# Test with valid request
curl -X POST http://localhost:8080/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"url":"https://httpbin.org/html"}'

# Check if target URL is accessible
curl -I https://example.com

# Verify request format
echo '{"url":"https://example.com"}' | jq .
```

#### 5. Memory Issues

**Symptoms**: Container killed by OOM or high memory usage

**Solutions**:
```bash
# Monitor memory usage
docker stats webailyzer-api

# Increase memory limit
docker run --memory="512m" webailyzer-lite-api

# Check for memory leaks in logs
docker logs webailyzer-api | grep -i memory
```

#### 6. Network Connectivity Issues

**Symptoms**: Cannot fetch external URLs

**Solutions**:
```bash
# Test network connectivity from container
docker exec webailyzer-api curl -I https://google.com

# Check DNS resolution
docker exec webailyzer-api nslookup google.com

# Verify firewall/proxy settings
# Check corporate proxy requirements
```

### Performance Tuning

#### 1. Concurrent Request Handling

The Go HTTP server handles concurrent requests automatically. For high load:

```bash
# Monitor concurrent connections
ss -tuln | grep :8080

# Use load testing tools
ab -n 1000 -c 10 http://localhost:8080/health
```

#### 2. Resource Optimization

```bash
# Monitor resource usage
docker stats --no-stream webailyzer-api

# Optimize for memory-constrained environments
docker run --memory="128m" --oom-kill-disable=false webailyzer-lite-api
```

### Getting Help

1. **Check logs first**:
   ```bash
   docker logs webailyzer-api
   ```

2. **Verify basic connectivity**:
   ```bash
   curl -v http://localhost:8080/health
   ```

3. **Test with simple request**:
   ```bash
   curl -X POST http://localhost:8080/v1/analyze \
     -H "Content-Type: application/json" \
     -d '{"url":"https://httpbin.org/html"}'
   ```

4. **Check system resources**:
   ```bash
   docker system df
   docker system prune
   ```

For additional support, please check the project's GitHub issues or create a new issue with:
- Error logs
- System information (OS, Docker version)
- Steps to reproduce the problem
- Expected vs actual behavior