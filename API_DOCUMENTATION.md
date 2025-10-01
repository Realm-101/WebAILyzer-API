# WebAIlyzer Lite API Documentation

## Overview

The WebAIlyzer Lite API provides comprehensive website analysis capabilities including performance metrics, SEO analysis, accessibility checks, security assessments, and AI-powered insights.

**Base URL:** `https://api.webailyzer.com`  
**Version:** v1  
**Authentication:** Bearer Token (API Key)

## Authentication

All API requests require authentication using a Bearer token in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Rate Limiting

API requests are rate-limited per workspace. Rate limit information is included in response headers:

- `X-RateLimit-Limit`: Maximum requests allowed in the time window
- `X-RateLimit-Remaining`: Remaining requests in the current window
- `X-RateLimit-Reset`: Time when the rate limit resets (Unix timestamp)
- `Retry-After`: Seconds to wait before retrying (when rate limited)

## Error Handling

The API uses standard HTTP status codes and returns error details in JSON format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error context"
  }
}
```

### Common Error Codes

- `INVALID_API_KEY`: Invalid or missing API key
- `WORKSPACE_INACTIVE`: Workspace is not active
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded
- `INVALID_REQUEST`: Invalid request parameters
- `RESOURCE_NOT_FOUND`: Requested resource not found
- `INTERNAL_ERROR`: Internal server error

## Endpoints

### Health Check

#### GET /health

Check API health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### Analysis Endpoints

#### POST /api/v1/analyze

Analyze a single URL for various metrics.

**Request Body:**
```json
{
  "url": "https://example.com",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "session_id": "550e8400-e29b-41d4-a716-446655440001",
  "options": {
    "include_performance": true,
    "include_seo": true,
    "include_accessibility": true,
    "include_security": true,
    "include_technologies": true
  }
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "session_id": "550e8400-e29b-41d4-a716-446655440001",
  "url": "https://example.com",
  "status": "completed",
  "technologies": {
    "detected": ["nginx", "php", "mysql"],
    "categories": {
      "web_servers": ["nginx"],
      "programming_languages": ["php"],
      "databases": ["mysql"]
    }
  },
  "performance_metrics": {
    "load_time_ms": 1250,
    "first_contentful_paint": 800,
    "largest_contentful_paint": 1200,
    "cumulative_layout_shift": 0.05,
    "time_to_interactive": 1500
  },
  "seo_metrics": {
    "title": "Example Website",
    "meta_description": "This is an example website",
    "h1_count": 1,
    "h2_count": 3,
    "image_alt_missing": 2,
    "internal_links": 15,
    "external_links": 5
  },
  "accessibility_metrics": {
    "score": 85,
    "violations": [
      {
        "rule": "color-contrast",
        "impact": "serious",
        "description": "Elements must have sufficient color contrast"
      }
    ],
    "passes": 12,
    "violations_count": 3
  },
  "security_metrics": {
    "https": true,
    "hsts": false,
    "ssl_grade": "A",
    "vulnerabilities": [],
    "security_headers": {
      "content_security_policy": false,
      "x_frame_options": true,
      "x_content_type_options": true
    }
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:05Z"
}
```

#### POST /api/v1/batch

Analyze multiple URLs in a single batch request.

**Request Body:**
```json
{
  "urls": [
    "https://example1.com",
    "https://example2.com",
    "https://example3.com"
  ],
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "session_id": "550e8400-e29b-41d4-a716-446655440001",
  "options": {
    "include_performance": true,
    "include_seo": true
  }
}
```

**Response:**
```json
{
  "batch_id": "550e8400-e29b-41d4-a716-446655440003",
  "status": "completed",
  "results": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "url": "https://example1.com",
      "status": "completed",
      "technologies": {...},
      "performance_metrics": {...}
    }
  ],
  "failed_urls": [
    "https://unreachable-site.com"
  ],
  "progress": {
    "total": 3,
    "completed": 2,
    "failed": 1
  },
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:32:00Z"
}
```

#### GET /api/v1/analysis

Retrieve analysis history for a workspace.

**Query Parameters:**
- `workspace_id` (required): Workspace UUID
- `limit` (optional): Number of results (default: 50, max: 1000)
- `offset` (optional): Pagination offset (default: 0)
- `start_date` (optional): Filter by start date (ISO 8601)
- `end_date` (optional): Filter by end date (ISO 8601)
- `url` (optional): Filter by URL pattern

**Response:**
```json
{
  "results": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "url": "https://example.com",
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "metadata": {
    "count": 1,
    "total": 150,
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "filters": {
      "limit": 50,
      "offset": 0
    }
  }
}
```

### Insights Endpoints

#### POST /api/v1/insights/generate

Generate AI-powered insights for a workspace.

**Request Body:**
```json
{
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "analysis_ids": ["550e8400-e29b-41d4-a716-446655440002"],
  "insight_types": ["performance", "seo", "accessibility"]
}
```

**Response:**
```json
{
  "success": true,
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "insights_generated": 5,
  "job_id": "550e8400-e29b-41d4-a716-446655440005"
}
```

#### GET /api/v1/insights

Retrieve insights for a workspace.

**Query Parameters:**
- `workspace_id` (required): Workspace UUID
- `status` (optional): Filter by status (pending, applied, dismissed)
- `priority` (optional): Filter by priority (low, medium, high, critical)
- `insight_type` (optional): Filter by type
- `limit` (optional): Number of results (default: 50)
- `offset` (optional): Pagination offset

**Response:**
```json
{
  "insights": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440006",
      "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
      "insight_type": "performance_bottleneck",
      "priority": "high",
      "title": "Optimize Image Loading",
      "description": "Large images are slowing down page load times",
      "impact_score": 85,
      "effort_score": 30,
      "recommendations": {
        "actions": [
          "Compress images using WebP format",
          "Implement lazy loading for below-fold images",
          "Use responsive image sizes"
        ],
        "expected_improvement": "40% faster load times"
      },
      "data_source": {
        "analysis_id": "550e8400-e29b-41d4-a716-446655440002",
        "url": "https://example.com"
      },
      "status": "pending",
      "created_at": "2024-01-15T10:35:00Z"
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "count": 1,
    "total": 12
  }
}
```

#### PUT /api/v1/insights/{insight_id}/status

Update the status of an insight.

**Request Body:**
```json
{
  "status": "applied",
  "notes": "Implemented image compression and lazy loading"
}
```

**Response:**
```json
{
  "success": true,
  "insight_id": "550e8400-e29b-41d4-a716-446655440006",
  "status": "applied",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

### Metrics Endpoints

#### GET /api/v1/metrics

Retrieve aggregated metrics for a workspace.

**Query Parameters:**
- `workspace_id` (required): Workspace UUID
- `start_date` (required): Start date (ISO 8601)
- `end_date` (required): End date (ISO 8601)
- `granularity` (required): Data granularity (hourly, daily, weekly, monthly)
- `metrics` (optional): Comma-separated list of specific metrics

**Response:**
```json
{
  "metrics": {
    "avg_load_time": {
      "current": 1850.5,
      "previous": 2100.2,
      "trend": "down",
      "data_points": [
        {
          "timestamp": "2024-01-15T00:00:00Z",
          "value": 1850.5
        }
      ]
    },
    "conversion_rate": {
      "current": 3.2,
      "previous": 2.8,
      "trend": "up",
      "data_points": [...]
    }
  },
  "kpis": [
    {
      "name": "Average Load Time",
      "value": 1850.5,
      "target": 2000.0,
      "status": "good",
      "description": "Average page load time in milliseconds"
    }
  ],
  "anomalies": [
    {
      "metric": "conversion_rate",
      "timestamp": "2024-01-15T10:00:00Z",
      "expected": 2.9,
      "actual": 3.2,
      "severity": "medium",
      "description": "Conversion rate spike detected"
    }
  ],
  "metadata": {
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "date_range": {
      "start": "2024-01-08T00:00:00Z",
      "end": "2024-01-15T00:00:00Z"
    },
    "granularity": "daily",
    "from_cache": false,
    "data_source": "real_time"
  }
}
```

### Export Endpoints

#### POST /api/v1/export

Export analysis data in various formats.

**Request Body:**
```json
{
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "format": "pdf",
  "data_types": ["analysis", "insights", "metrics"],
  "filters": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-15T23:59:59Z",
    "urls": ["https://example.com"]
  },
  "options": {
    "include_charts": true,
    "include_recommendations": true
  }
}
```

**Response:**
```json
{
  "export_id": "550e8400-e29b-41d4-a716-446655440007",
  "status": "processing",
  "format": "pdf",
  "estimated_completion": "2024-01-15T10:45:00Z",
  "download_url": null
}
```

#### GET /api/v1/export/{export_id}

Check export status and download when ready.

**Response:**
```json
{
  "export_id": "550e8400-e29b-41d4-a716-446655440007",
  "status": "completed",
  "format": "pdf",
  "file_size": 2048576,
  "download_url": "https://api.webailyzer.com/downloads/export_123.pdf",
  "expires_at": "2024-01-22T10:45:00Z",
  "created_at": "2024-01-15T10:40:00Z",
  "completed_at": "2024-01-15T10:44:30Z"
}
```

## Code Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://api.webailyzer.com',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});

// Analyze a URL
async function analyzeURL(url, workspaceId) {
  try {
    const response = await client.post('/api/v1/analyze', {
      url: url,
      workspace_id: workspaceId,
      options: {
        include_performance: true,
        include_seo: true,
        include_accessibility: true,
        include_security: true
      }
    });
    
    console.log('Analysis completed:', response.data);
    return response.data;
  } catch (error) {
    console.error('Analysis failed:', error.response.data);
    throw error;
  }
}

// Get insights
async function getInsights(workspaceId) {
  try {
    const response = await client.get('/api/v1/insights', {
      params: {
        workspace_id: workspaceId,
        status: 'pending',
        limit: 10
      }
    });
    
    return response.data.insights;
  } catch (error) {
    console.error('Failed to get insights:', error.response.data);
    throw error;
  }
}
```

### Python

```python
import requests
import json

class WebAIlyzerClient:
    def __init__(self, api_key, base_url='https://api.webailyzer.com'):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }
    
    def analyze_url(self, url, workspace_id, options=None):
        """Analyze a single URL"""
        if options is None:
            options = {
                'include_performance': True,
                'include_seo': True,
                'include_accessibility': True,
                'include_security': True
            }
        
        payload = {
            'url': url,
            'workspace_id': workspace_id,
            'options': options
        }
        
        response = requests.post(
            f'{self.base_url}/api/v1/analyze',
            headers=self.headers,
            json=payload
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception(f'Analysis failed: {response.json()}')
    
    def batch_analyze(self, urls, workspace_id, options=None):
        """Analyze multiple URLs"""
        payload = {
            'urls': urls,
            'workspace_id': workspace_id,
            'options': options or {}
        }
        
        response = requests.post(
            f'{self.base_url}/api/v1/batch',
            headers=self.headers,
            json=payload
        )
        
        return response.json()
    
    def get_insights(self, workspace_id, status=None, priority=None):
        """Get insights for a workspace"""
        params = {'workspace_id': workspace_id}
        if status:
            params['status'] = status
        if priority:
            params['priority'] = priority
        
        response = requests.get(
            f'{self.base_url}/api/v1/insights',
            headers=self.headers,
            params=params
        )
        
        return response.json()

# Usage example
client = WebAIlyzerClient('your-api-key')

# Analyze a URL
result = client.analyze_url(
    'https://example.com',
    'your-workspace-id'
)

print(f"Analysis completed: {result['status']}")
print(f"Load time: {result['performance_metrics']['load_time_ms']}ms")
```

### cURL Examples

```bash
# Analyze a URL
curl -X POST https://api.webailyzer.com/api/v1/analyze \
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

# Get insights
curl -X GET "https://api.webailyzer.com/api/v1/insights?workspace_id=550e8400-e29b-41d4-a716-446655440000&status=pending" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Generate insights
curl -X POST https://api.webailyzer.com/api/v1/insights/generate \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# Get metrics
curl -X GET "https://api.webailyzer.com/api/v1/metrics?workspace_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-15T23:59:59Z&granularity=daily" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Webhooks

The API supports webhooks for real-time notifications of analysis completion and insight generation.

### Webhook Events

- `analysis.completed`: Fired when an analysis is completed
- `batch.completed`: Fired when a batch analysis is completed
- `insights.generated`: Fired when new insights are generated
- `export.completed`: Fired when an export is ready for download

### Webhook Payload Example

```json
{
  "event": "analysis.completed",
  "timestamp": "2024-01-15T10:30:05Z",
  "data": {
    "analysis_id": "550e8400-e29b-41d4-a716-446655440002",
    "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
    "url": "https://example.com",
    "status": "completed"
  }
}
```

## SDKs and Libraries

Official SDKs are available for:

- **JavaScript/Node.js**: `npm install @webailyzer/sdk`
- **Python**: `pip install webailyzer-sdk`
- **PHP**: `composer require webailyzer/sdk`
- **Go**: `go get github.com/webailyzer/go-sdk`

## Support

- **Documentation**: https://docs.webailyzer.com
- **API Status**: https://status.webailyzer.com
- **Support Email**: api-support@webailyzer.com
- **GitHub Issues**: https://github.com/webailyzer/api-issues
---


# Deployment and Configuration Guide

## Environment Variables

The following environment variables are required for deployment:

### Database Configuration
```bash
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=webailyzer
DATABASE_PASSWORD=your_secure_password
DATABASE_NAME=webailyzer
DATABASE_SSL_MODE=require
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5
```

### Redis Configuration
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
REDIS_MAX_CONNECTIONS=10
```

### Application Configuration
```bash
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=production
API_BASE_URL=https://api.webailyzer.com
CORS_ALLOWED_ORIGINS=https://app.webailyzer.com,https://dashboard.webailyzer.com
```

### Rate Limiting
```bash
RATE_LIMIT_DEFAULT=1000
RATE_LIMIT_WINDOW_DURATION=1h
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

### Monitoring and Observability
```bash
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
SENTRY_DSN=your_sentry_dsn
```

## Docker Deployment

### Using Docker Compose

1. **Create docker-compose.yml:**

```yaml
version: '3.8'

services:
  webailyzer-api:
    image: webailyzer/api:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=webailyzer
      - DATABASE_PASSWORD=secure_password
      - DATABASE_NAME=webailyzer
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LOG_LEVEL=info
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: webailyzer
      POSTGRES_PASSWORD: secure_password
      POSTGRES_DB: webailyzer
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U webailyzer"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass redis_password
    volumes:
      - redis_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "auth", "redis_password", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - webailyzer-api
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

2. **Start the services:**

```bash
docker-compose up -d
```

### Kubernetes Deployment

1. **Create namespace:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: webailyzer
```

2. **ConfigMap for environment variables:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webailyzer-config
  namespace: webailyzer
data:
  DATABASE_HOST: "postgres-service"
  DATABASE_PORT: "5432"
  DATABASE_NAME: "webailyzer"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
```

3. **Secret for sensitive data:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: webailyzer-secrets
  namespace: webailyzer
type: Opaque
data:
  DATABASE_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
```

4. **Deployment:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webailyzer-api
  namespace: webailyzer
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webailyzer-api
  template:
    metadata:
      labels:
        app: webailyzer-api
    spec:
      containers:
      - name: webailyzer-api
        image: webailyzer/api:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: webailyzer-config
        - secretRef:
            name: webailyzer-secrets
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

5. **Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: webailyzer-api-service
  namespace: webailyzer
spec:
  selector:
    app: webailyzer-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

## Database Setup

### PostgreSQL Setup

1. **Create database and user:**

```sql
CREATE DATABASE webailyzer;
CREATE USER webailyzer WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE webailyzer TO webailyzer;
```

2. **Run migrations:**

```bash
# Using migrate tool
migrate -path ./internal/database/migrations -database "postgres://webailyzer:password@localhost:5432/webailyzer?sslmode=disable" up

# Or using the application
./webailyzer-api migrate
```

### Database Performance Tuning

Add these settings to your PostgreSQL configuration:

```ini
# postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
```

## Load Balancing and High Availability

### Nginx Configuration

```nginx
upstream webailyzer_backend {
    least_conn;
    server webailyzer-api-1:8080 max_fails=3 fail_timeout=30s;
    server webailyzer-api-2:8080 max_fails=3 fail_timeout=30s;
    server webailyzer-api-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.webailyzer.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.webailyzer.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;

    location / {
        proxy_pass http://webailyzer_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Rate limiting
        limit_req zone=api burst=20 nodelay;
    }

    location /health {
        proxy_pass http://webailyzer_backend;
        access_log off;
    }
}

# Rate limiting zone
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
}
```

## Monitoring and Observability

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'webailyzer-api'
    static_configs:
      - targets: ['webailyzer-api:8080']
    metrics_path: /metrics
    scrape_interval: 30s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Grafana Dashboard

Import the dashboard from `monitoring/grafana-dashboard.json` or create custom dashboards with these key metrics:

- Request rate and latency
- Error rates by endpoint
- Database connection pool usage
- Redis cache hit/miss rates
- Memory and CPU usage
- Analysis processing times

## Security Configuration

### SSL/TLS Setup

1. **Generate SSL certificates:**

```bash
# Using Let's Encrypt
certbot certonly --webroot -w /var/www/html -d api.webailyzer.com

# Or use your own certificates
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/webailyzer.key \
  -out /etc/ssl/certs/webailyzer.crt
```

2. **Configure security headers:**

```nginx
# Add to nginx server block
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Content-Type-Options nosniff;
add_header X-Frame-Options DENY;
add_header X-XSS-Protection "1; mode=block";
add_header Referrer-Policy "strict-origin-when-cross-origin";
```

### Firewall Configuration

```bash
# UFW (Ubuntu)
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw deny 8080/tcp   # Block direct API access
ufw enable

# iptables
iptables -A INPUT -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

## Backup and Recovery

### Database Backup

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="webailyzer"

# Create backup
pg_dump -h localhost -U webailyzer -d $DB_NAME | gzip > $BACKUP_DIR/webailyzer_$DATE.sql.gz

# Keep only last 7 days of backups
find $BACKUP_DIR -name "webailyzer_*.sql.gz" -mtime +7 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/webailyzer_$DATE.sql.gz s3://webailyzer-backups/
```

### Automated Backup with Cron

```bash
# Add to crontab
0 2 * * * /path/to/backup.sh
```

### Recovery Process

```bash
# Restore from backup
gunzip -c /backups/webailyzer_20240115_020000.sql.gz | psql -h localhost -U webailyzer -d webailyzer
```

## Performance Optimization

### Application-Level Optimizations

1. **Connection Pooling:**
   - Database: 25 max connections, 5 idle
   - Redis: 10 max connections

2. **Caching Strategy:**
   - Analysis results: 1 hour TTL
   - Metrics data: 15 minutes TTL
   - Insights: 30 minutes TTL

3. **Request Timeouts:**
   - Analysis requests: 60 seconds
   - Batch requests: 300 seconds
   - Database queries: 30 seconds

### Infrastructure Optimizations

1. **Database:**
   - Use read replicas for analytics queries
   - Implement connection pooling (PgBouncer)
   - Regular VACUUM and ANALYZE

2. **Caching:**
   - Redis cluster for high availability
   - CDN for static assets
   - Application-level caching

3. **Load Balancing:**
   - Multiple API instances
   - Health check endpoints
   - Circuit breaker pattern

## Troubleshooting Guide

### Common Issues

1. **Database Connection Issues:**
   ```bash
   # Check database connectivity
   pg_isready -h localhost -p 5432 -U webailyzer
   
   # Check connection pool status
   curl http://localhost:8080/debug/db/stats
   ```

2. **Redis Connection Issues:**
   ```bash
   # Test Redis connectivity
   redis-cli -h localhost -p 6379 ping
   
   # Check Redis memory usage
   redis-cli info memory
   ```

3. **High Memory Usage:**
   ```bash
   # Check application memory usage
   curl http://localhost:8080/debug/pprof/heap
   
   # Monitor with top/htop
   top -p $(pgrep webailyzer-api)
   ```

4. **Slow API Responses:**
   ```bash
   # Check slow queries
   curl http://localhost:8080/debug/db/slow-queries
   
   # Monitor request latency
   curl http://localhost:8080/metrics | grep http_request_duration
   ```

### Log Analysis

```bash
# View application logs
docker logs webailyzer-api

# Filter error logs
docker logs webailyzer-api 2>&1 | grep ERROR

# Monitor real-time logs
docker logs -f webailyzer-api
```

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Database health
curl http://localhost:8080/health/db

# Redis health
curl http://localhost:8080/health/redis
```

## Scaling Considerations

### Horizontal Scaling

1. **Stateless Design:** The API is designed to be stateless, allowing easy horizontal scaling
2. **Load Balancing:** Use nginx or cloud load balancers to distribute traffic
3. **Database Scaling:** Implement read replicas and connection pooling
4. **Cache Scaling:** Use Redis cluster for distributed caching

### Vertical Scaling

1. **CPU:** Analysis operations are CPU-intensive
2. **Memory:** Batch processing requires adequate memory
3. **Storage:** Database and log storage requirements
4. **Network:** High throughput for API requests

### Auto-scaling Configuration

```yaml
# Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: webailyzer-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: webailyzer-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```