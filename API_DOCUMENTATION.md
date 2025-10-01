# WebAIlyzer Lite API Documentation

## Overview

The WebAIlyzer Lite API is a simple web technology detection service that analyzes websites to identify the technologies, frameworks, and libraries they use.

**Base URL:** `http://localhost:8080`  
**Version:** v1  
**Authentication:** None required

## Endpoints

## Error Handling

The API uses standard HTTP status codes and returns error details in JSON format:

```json
{
  "error": "Human-readable error message"
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid JSON or missing required fields
- `502 Bad Gateway`: Failed to fetch the provided URL
- `500 Internal Server Error`: Wappalyzer engine initialization failed

### Health Check

#### GET /health

Check if the API service is running and healthy.

**Response:**
```json
{
  "status": "ok"
}
```

**Status Codes:**
- `200 OK`: Service is healthy

### Website Analysis

#### POST /v1/analyze

Analyze a website to detect technologies, frameworks, and libraries.

**Request Body:**
```json
{
  "url": "https://example.com"
}
```

**Parameters:**
- `url` (string, required): The URL of the website to analyze

**Response:**
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
    },
    "jQuery": {
      "categories": ["JavaScript libraries"],
      "confidence": 100,
      "version": "3.4.1",
      "icon": "jQuery.svg",
      "website": "https://jquery.com"
    }
  },
  "content_type": "text/html; charset=utf-8"
}
```

**Response Fields:**
- `url`: The analyzed URL
- `detected`: Object containing detected technologies with their details
- `content_type`: The content type of the analyzed page

**Status Codes:**
- `200 OK`: Analysis completed successfully
- `400 Bad Request`: Invalid JSON or missing URL field
- `502 Bad Gateway`: Failed to fetch the provided URL
- `500 Internal Server Error`: Wappalyzer engine initialization failed



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