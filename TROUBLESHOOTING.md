# WebAIlyzer API Troubleshooting Guide

## Quick Diagnostics

### Health Check Commands

```bash
# Basic API health
curl -f http://localhost:8080/health

# Detailed health with component status
curl http://localhost:8080/health/detailed

# Database connectivity
curl http://localhost:8080/health/db

# Redis connectivity  
curl http://localhost:8080/health/redis

# Check API version and build info
curl http://localhost:8080/version
```

### Log Analysis

```bash
# View recent logs
docker logs --tail 100 webailyzer-api

# Follow logs in real-time
docker logs -f webailyzer-api

# Filter error logs
docker logs webailyzer-api 2>&1 | grep -E "(ERROR|FATAL|PANIC)"

# Search for specific patterns
docker logs webailyzer-api 2>&1 | grep "database"
```

## Common Issues and Solutions

### 1. API Not Starting

**Symptoms:**
- Container exits immediately
- "Connection refused" errors
- Health check failures

**Diagnostic Steps:**
```bash
# Check container status
docker ps -a

# View startup logs
docker logs webailyzer-api

# Check environment variables
docker exec webailyzer-api env | grep -E "(DATABASE|REDIS)"
```

**Common Causes & Solutions:**

**Missing Environment Variables:**
```bash
# Verify required variables are set
echo $DATABASE_HOST
echo $DATABASE_PASSWORD
echo $REDIS_HOST

# Set missing variables
export DATABASE_HOST=localhost
export DATABASE_PASSWORD=your_password
```

**Database Connection Issues:**
```bash
# Test database connectivity
pg_isready -h $DATABASE_HOST -p $DATABASE_PORT -U $DATABASE_USER

# Check database exists
psql -h $DATABASE_HOST -U $DATABASE_USER -l | grep webailyzer

# Run migrations if needed
./webailyzer-api migrate up
```

**Port Conflicts:**
```bash
# Check if port is in use
netstat -tulpn | grep :8080
lsof -i :8080

# Use different port
export PORT=8081
```

### 2. Database Connection Problems

**Symptoms:**
- "connection refused" errors
- "too many connections" errors
- Slow query performance

**Diagnostic Commands:**
```bash
# Check database connectivity
pg_isready -h $DATABASE_HOST -p $DATABASE_PORT -U $DATABASE_USER

# Check active connections
psql -h $DATABASE_HOST -U $DATABASE_USER -c "SELECT count(*) FROM pg_stat_activity;"

# Check connection pool status
curl http://localhost:8080/debug/db/stats
```

**Solutions:**

**Connection Pool Exhaustion:**
```bash
# Increase max connections in config
export DATABASE_MAX_CONNECTIONS=50
export DATABASE_MAX_IDLE_CONNECTIONS=10

# Check for connection leaks
curl http://localhost:8080/debug/db/stats | jq '.open_connections'
```

**Slow Queries:**
```sql
-- Find slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Check for missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation 
FROM pg_stats 
WHERE schemaname = 'public';
```

**Database Lock Issues:**
```sql
-- Check for locks
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

### 3. Redis Connection Issues

**Symptoms:**
- Cache misses
- "connection refused" errors
- Slow response times

**Diagnostic Commands:**
```bash
# Test Redis connectivity
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping

# Check Redis info
redis-cli -h $REDIS_HOST -p $REDIS_PORT info

# Monitor Redis commands
redis-cli -h $REDIS_HOST -p $REDIS_PORT monitor
```

**Solutions:**

**Connection Issues:**
```bash
# Check Redis is running
systemctl status redis
docker ps | grep redis

# Test authentication
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping

# Check network connectivity
telnet $REDIS_HOST $REDIS_PORT
```

**Memory Issues:**
```bash
# Check Redis memory usage
redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory

# Clear cache if needed
redis-cli -h $REDIS_HOST -p $REDIS_PORT flushdb

# Set memory limit
redis-cli -h $REDIS_HOST -p $REDIS_PORT config set maxmemory 1gb
```

### 4. High Memory Usage

**Symptoms:**
- Out of memory errors
- Slow performance
- Container restarts

**Diagnostic Commands:**
```bash
# Check container memory usage
docker stats webailyzer-api

# Get memory profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Check system memory
free -h
top -p $(pgrep webailyzer-api)
```

**Solutions:**

**Memory Leaks:**
```bash
# Enable memory profiling
export ENABLE_PPROF=true

# Monitor memory over time
while true; do
  curl -s http://localhost:8080/debug/pprof/heap | head -1
  sleep 60
done

# Restart container if needed
docker restart webailyzer-api
```

**Optimize Configuration:**
```bash
# Reduce batch sizes
export MAX_BATCH_SIZE=25

# Limit concurrent analyses
export MAX_CONCURRENT_ANALYSES=5

# Increase garbage collection frequency
export GOGC=50
```

### 5. Slow API Performance

**Symptoms:**
- High response times
- Timeouts
- Rate limit errors

**Diagnostic Commands:**
```bash
# Check response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/analyze

# Monitor metrics
curl http://localhost:8080/metrics | grep http_request_duration

# Check database performance
curl http://localhost:8080/debug/db/slow-queries
```

**Solutions:**

**Database Optimization:**
```sql
-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_analysis_workspace_created 
ON analysis(workspace_id, created_at);

-- Update statistics
ANALYZE;

-- Vacuum if needed
VACUUM ANALYZE analysis;
```

**Cache Optimization:**
```bash
# Increase cache TTL
export CACHE_TTL=3600

# Pre-warm cache
curl -X POST http://localhost:8080/api/v1/cache/warm

# Check cache hit rate
redis-cli -h $REDIS_HOST -p $REDIS_PORT info stats | grep keyspace
```

**Connection Tuning:**
```bash
# Increase timeouts
export HTTP_READ_TIMEOUT=60s
export HTTP_WRITE_TIMEOUT=60s

# Optimize connection pools
export DATABASE_MAX_CONNECTIONS=25
export REDIS_MAX_CONNECTIONS=10
```

### 6. Rate Limiting Issues

**Symptoms:**
- 429 Too Many Requests errors
- Blocked API calls
- Inconsistent rate limits

**Diagnostic Commands:**
```bash
# Check rate limit headers
curl -I http://localhost:8080/api/v1/analyze

# Monitor rate limit metrics
curl http://localhost:8080/metrics | grep rate_limit

# Check workspace rate limits
curl http://localhost:8080/api/v1/workspaces/{id}/limits
```

**Solutions:**

**Adjust Rate Limits:**
```bash
# Increase default limits
export RATE_LIMIT_DEFAULT=2000
export RATE_LIMIT_WINDOW_DURATION=1h

# Per-workspace limits
curl -X PUT http://localhost:8080/api/v1/workspaces/{id}/limits \
  -d '{"rate_limit": 5000}'
```

**Rate Limit Bypass:**
```bash
# Temporary bypass for testing
export RATE_LIMIT_ENABLED=false

# Whitelist specific IPs
export RATE_LIMIT_WHITELIST="192.168.1.0/24,10.0.0.0/8"
```

### 7. Analysis Failures

**Symptoms:**
- Analysis timeouts
- Invalid results
- Partial analysis data

**Diagnostic Commands:**
```bash
# Check analysis logs
docker logs webailyzer-api 2>&1 | grep "analysis"

# Monitor analysis metrics
curl http://localhost:8080/metrics | grep analysis_duration

# Check failed analyses
curl http://localhost:8080/api/v1/analysis/failed
```

**Solutions:**

**Timeout Issues:**
```bash
# Increase analysis timeout
export ANALYSIS_TIMEOUT=120s

# Reduce concurrent analyses
export MAX_CONCURRENT_ANALYSES=3

# Skip heavy analysis components
export SKIP_ACCESSIBILITY_CHECK=true
```

**Network Issues:**
```bash
# Test URL accessibility
curl -I https://example.com

# Check DNS resolution
nslookup example.com

# Use custom user agent
export USER_AGENT="WebAIlyzer/1.0"
```

**SSL/TLS Issues:**
```bash
# Skip SSL verification (not recommended for production)
export SKIP_SSL_VERIFY=true

# Update CA certificates
apt-get update && apt-get install ca-certificates
```

## Performance Monitoring

### Key Metrics to Monitor

```bash
# API Response Times
curl http://localhost:8080/metrics | grep http_request_duration_seconds

# Database Performance
curl http://localhost:8080/metrics | grep database_query_duration_seconds

# Memory Usage
curl http://localhost:8080/metrics | grep go_memstats_alloc_bytes

# Cache Hit Rate
curl http://localhost:8080/metrics | grep cache_hit_ratio

# Active Connections
curl http://localhost:8080/metrics | grep database_connections_active
```

### Setting Up Alerts

**Prometheus Alert Rules:**
```yaml
groups:
- name: webailyzer-api
  rules:
  - alert: HighResponseTime
    expr: histogram_quantile(0.95, http_request_duration_seconds_bucket) > 5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High API response time"

  - alert: DatabaseConnectionsHigh
    expr: database_connections_active / database_connections_max > 0.8
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Database connection pool nearly exhausted"

  - alert: HighMemoryUsage
    expr: go_memstats_alloc_bytes / 1024 / 1024 > 512
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage detected"
```

## Debug Mode

### Enabling Debug Features

```bash
# Enable debug mode
export DEBUG=true
export LOG_LEVEL=debug

# Enable profiling endpoints
export ENABLE_PPROF=true

# Enable debug routes
export ENABLE_DEBUG_ROUTES=true
```

### Debug Endpoints

```bash
# CPU profiling
curl http://localhost:8080/debug/pprof/profile > cpu.prof

# Memory profiling
curl http://localhost:8080/debug/pprof/heap > heap.prof

# Goroutine dump
curl http://localhost:8080/debug/pprof/goroutine > goroutines.txt

# Database statistics
curl http://localhost:8080/debug/db/stats

# Cache statistics
curl http://localhost:8080/debug/cache/stats

# Configuration dump
curl http://localhost:8080/debug/config
```

## Recovery Procedures

### Database Recovery

```bash
# Stop API
docker stop webailyzer-api

# Restore from backup
gunzip -c backup.sql.gz | psql -h $DATABASE_HOST -U $DATABASE_USER -d $DATABASE_NAME

# Run migrations
./webailyzer-api migrate up

# Start API
docker start webailyzer-api
```

### Cache Recovery

```bash
# Clear corrupted cache
redis-cli -h $REDIS_HOST -p $REDIS_PORT flushall

# Restart Redis
docker restart redis

# Warm cache
curl -X POST http://localhost:8080/api/v1/cache/warm
```

### Full System Recovery

```bash
# Stop all services
docker-compose down

# Restore data
./restore-backup.sh

# Start services
docker-compose up -d

# Verify health
curl http://localhost:8080/health
```

## Getting Help

### Log Collection

```bash
# Collect all relevant logs
mkdir -p debug-logs
docker logs webailyzer-api > debug-logs/api.log
docker logs postgres > debug-logs/postgres.log
docker logs redis > debug-logs/redis.log

# System information
uname -a > debug-logs/system.txt
docker version > debug-logs/docker.txt
free -h > debug-logs/memory.txt
df -h > debug-logs/disk.txt

# Create archive
tar -czf webailyzer-debug-$(date +%Y%m%d).tar.gz debug-logs/
```

### Support Information

When contacting support, please include:

1. **Error Description:** What you were trying to do and what happened
2. **Environment:** OS, Docker version, deployment method
3. **Configuration:** Relevant environment variables (sanitized)
4. **Logs:** Recent application and system logs
5. **Metrics:** Current system metrics and performance data
6. **Timeline:** When the issue started and any recent changes

### Contact Information

- **GitHub Issues:** https://github.com/webailyzer/api/issues
- **Support Email:** support@webailyzer.com
- **Documentation:** https://docs.webailyzer.com
- **Status Page:** https://status.webailyzer.com