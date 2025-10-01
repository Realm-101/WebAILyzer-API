# WebAIlyzer Lite API Monitoring

This directory contains monitoring configuration files for the WebAIlyzer Lite API.

## Metrics Collected

The API exposes Prometheus metrics on the `/metrics` endpoint. The following metrics are collected:

### HTTP Metrics
- `webailyzer_http_requests_total` - Total number of HTTP requests
- `webailyzer_http_request_duration_seconds` - Duration of HTTP requests
- `webailyzer_http_active_connections` - Number of active HTTP connections

### Analysis Metrics
- `webailyzer_analysis_operations_total` - Total number of analysis operations
- `webailyzer_analysis_duration_seconds` - Duration of analysis operations
- `webailyzer_analysis_errors_total` - Total number of analysis errors

### Database Metrics
- `webailyzer_database_connections` - Number of active database connections
- `webailyzer_database_operations_total` - Total number of database operations
- `webailyzer_database_operation_duration_seconds` - Duration of database operations

### Cache Metrics
- `webailyzer_cache_operations_total` - Total number of cache operations
- `webailyzer_cache_hit_ratio` - Cache hit ratio (0-1)

### Business Metrics
- `webailyzer_workspaces_total` - Total number of active workspaces
- `webailyzer_sessions_total` - Total number of user sessions
- `webailyzer_insights_generated_total` - Total number of insights generated

## Configuration Files

### prometheus.yml
Prometheus configuration file that defines:
- Scrape targets for the WebAIlyzer API
- Alerting rules
- Global settings

### webailyzer_rules.yml
Alerting rules for monitoring the API health:
- High error rate alerts
- High response time alerts
- Database connection alerts
- Cache performance alerts
- Analysis error alerts
- Service availability alerts

### grafana-dashboard.json
Grafana dashboard configuration that provides:
- HTTP request rate and duration graphs
- Analysis operation metrics
- Database and cache performance
- Error rate monitoring
- Business metrics visualization

## Setup Instructions

### 1. Prometheus Setup

1. Install Prometheus
2. Copy `prometheus.yml` to your Prometheus configuration directory
3. Copy `webailyzer_rules.yml` to your Prometheus rules directory
4. Update the target address in `prometheus.yml` if needed
5. Restart Prometheus

### 2. Grafana Setup

1. Install Grafana
2. Add Prometheus as a data source
3. Import the dashboard using `grafana-dashboard.json`
4. Configure alerts if needed

### 3. Docker Compose Setup

You can also use Docker Compose to set up the monitoring stack:

```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/webailyzer_rules.yml:/etc/prometheus/webailyzer_rules.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
```

### 4. Accessing Metrics

- Prometheus metrics: `http://localhost:8080/metrics`
- Prometheus UI: `http://localhost:9090`
- Grafana UI: `http://localhost:3000` (admin/admin)

## Alert Configuration

The monitoring setup includes several pre-configured alerts:

- **HighErrorRate**: Triggers when 5xx error rate exceeds 0.1 requests/second
- **HighResponseTime**: Triggers when 95th percentile response time exceeds 5 seconds
- **DatabaseConnectionsHigh**: Triggers when database connections exceed 80
- **CacheHitRateLow**: Triggers when cache hit ratio falls below 50%
- **AnalysisErrorsHigh**: Triggers when analysis error rate exceeds 0.05 errors/second
- **ServiceDown**: Triggers when the service is unreachable
- **SlowAnalysisOperations**: Triggers when analysis operations take longer than 30 seconds

## Customization

You can customize the monitoring setup by:

1. Modifying alert thresholds in `webailyzer_rules.yml`
2. Adding new panels to the Grafana dashboard
3. Creating custom metrics in the application code
4. Adding additional scrape targets for related services

## Troubleshooting

### Common Issues

1. **Metrics not appearing**: Check that the `/metrics` endpoint is accessible
2. **Alerts not firing**: Verify Prometheus can reach the API and rules are loaded
3. **Dashboard not loading**: Ensure Grafana can connect to Prometheus data source
4. **High cardinality warnings**: Review metric labels to avoid too many unique combinations

### Debugging

- Check Prometheus targets: `http://localhost:9090/targets`
- Verify rules: `http://localhost:9090/rules`
- Test queries: Use Prometheus query interface
- Check logs: Review application and Prometheus logs for errors