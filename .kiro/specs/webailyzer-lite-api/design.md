# Design Document

## Overview

The WebAIlyzer Lite API transforms the existing wappalyzer technology detection server into a comprehensive web analytics platform. The design extends the current Go-based HTTP server with persistent storage, enhanced analysis capabilities, AI-powered insights, and robust API endpoints that support the full WebAIlyzer Lite ecosystem.

The system maintains the lightweight, self-hosted nature of the original while adding enterprise-grade features for analytics, session tracking, and intelligent recommendations.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │  Web Dashboard  │    │  External APIs  │
│                 │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   API Gateway   │
                    │  (Rate Limiting │
                    │ Authentication) │
                    └─────────┬───────┘
                              │
                 ┌─────────────────────┐
                 │   WebAIlyzer API    │
                 │                     │
                 │  ┌───────────────┐  │
                 │  │   Enhanced    │  │
                 │  │   Analysis    │  │
                 │  │   Engine      │  │
                 │  └───────────────┘  │
                 │                     │
                 │  ┌───────────────┐  │
                 │  │  AI Insights  │  │
                 │  │    Engine     │  │
                 │  └───────────────┘  │
                 └─────────┬───────────┘
                           │
              ┌─────────────────────┐
              │   Data Layer        │
              │                     │
              │ ┌─────────────────┐ │
              │ │   PostgreSQL    │ │
              │ │   (Primary)     │ │
              │ └─────────────────┘ │
              │                     │
              │ ┌─────────────────┐ │
              │ │     Redis       │ │
              │ │   (Caching)     │ │
              │ └─────────────────┘ │
              └─────────────────────┘
```

### Component Architecture

The system follows a layered architecture pattern:

1. **API Layer**: HTTP handlers with middleware for authentication, rate limiting, and request validation
2. **Service Layer**: Business logic for analysis, insights generation, and data processing
3. **Repository Layer**: Data access abstraction for database operations
4. **External Integration Layer**: Interfaces for third-party services and webhooks

## Components and Interfaces

### API Endpoints

#### Enhanced Analysis Endpoint
```go
POST /api/v1/analyze
Content-Type: application/json

{
  "url": "https://example.com",
  "session_id": "optional-session-uuid",
  "workspace_id": "workspace-uuid",
  "options": {
    "include_performance": true,
    "include_seo": true,
    "include_accessibility": true,
    "include_security": true,
    "user_agent": "custom-agent"
  }
}

Response:
{
  "analysis_id": "uuid",
  "url": "https://example.com",
  "timestamp": "2024-01-01T00:00:00Z",
  "session_id": "session-uuid",
  "technologies": {...},
  "performance": {
    "load_time_ms": 1250,
    "first_contentful_paint": 800,
    "largest_contentful_paint": 1100,
    "resource_count": 45,
    "total_size_kb": 2048
  },
  "seo": {
    "title": "Page Title",
    "meta_description": "Description",
    "h1_count": 1,
    "structured_data": [...],
    "score": 85
  },
  "accessibility": {
    "score": 78,
    "issues": [...],
    "wcag_level": "AA"
  },
  "security": {
    "https": true,
    "headers": {...},
    "score": 92
  }
}
```

#### Batch Analysis Endpoint
```go
POST /api/v1/batch
Content-Type: application/json

{
  "urls": ["https://example1.com", "https://example2.com"],
  "workspace_id": "workspace-uuid",
  "options": {...}
}

Response:
{
  "batch_id": "uuid",
  "status": "processing|completed|failed",
  "results": [...],
  "failed_urls": [...],
  "progress": {
    "completed": 8,
    "total": 10
  }
}
```

#### Insights Endpoint
```go
GET /api/v1/insights?workspace_id=uuid&limit=10&status=pending

Response:
{
  "insights": [
    {
      "id": "uuid",
      "type": "performance_bottleneck",
      "priority": "high",
      "title": "Mobile page load time exceeds 3 seconds",
      "description": "Your mobile pages are loading slowly...",
      "impact_score": 85,
      "effort_score": 30,
      "recommendations": [...],
      "data_source": {...},
      "created_at": "2024-01-01T00:00:00Z",
      "status": "pending|applied|dismissed"
    }
  ],
  "pagination": {...}
}
```

#### Events Endpoint
```go
POST /api/v1/events
Content-Type: application/json

{
  "session_id": "session-uuid",
  "workspace_id": "workspace-uuid",
  "events": [
    {
      "event_id": "uuid",
      "type": "pageview|click|conversion",
      "timestamp": "2024-01-01T00:00:00Z",
      "url": "https://example.com/page",
      "properties": {...}
    }
  ]
}
```

#### Metrics Endpoint
```go
GET /api/v1/metrics?workspace_id=uuid&start_date=2024-01-01&end_date=2024-01-31&granularity=daily

Response:
{
  "metrics": {
    "conversion_rate": {
      "current": 3.2,
      "previous": 2.8,
      "trend": "up",
      "data_points": [...]
    },
    "bounce_rate": {...},
    "avg_session_duration": {...},
    "page_load_time": {...}
  },
  "kpis": [...],
  "anomalies": [...]
}
```

### Core Services

#### AnalysisService
```go
type AnalysisService interface {
    AnalyzeURL(ctx context.Context, req *AnalysisRequest) (*AnalysisResult, error)
    BatchAnalyze(ctx context.Context, req *BatchAnalysisRequest) (*BatchAnalysisResult, error)
    GetAnalysisHistory(ctx context.Context, workspaceID string, filters *AnalysisFilters) ([]*AnalysisResult, error)
}

type AnalysisRequest struct {
    URL         string
    SessionID   *string
    WorkspaceID string
    Options     AnalysisOptions
}

type AnalysisOptions struct {
    IncludePerformance   bool
    IncludeSEO          bool
    IncludeAccessibility bool
    IncludeSecurity     bool
    UserAgent           string
}
```

#### InsightsService
```go
type InsightsService interface {
    GenerateInsights(ctx context.Context, workspaceID string) ([]*Insight, error)
    GetInsights(ctx context.Context, workspaceID string, filters *InsightFilters) ([]*Insight, error)
    UpdateInsightStatus(ctx context.Context, insightID string, status InsightStatus) error
}

type Insight struct {
    ID              string
    Type            InsightType
    Priority        Priority
    Title           string
    Description     string
    ImpactScore     int
    EffortScore     int
    Recommendations []Recommendation
    DataSource      map[string]interface{}
    CreatedAt       time.Time
    Status          InsightStatus
}
```

#### EventService
```go
type EventService interface {
    TrackEvents(ctx context.Context, req *EventTrackingRequest) error
    GetEvents(ctx context.Context, filters *EventFilters) ([]*Event, error)
    GetSessions(ctx context.Context, filters *SessionFilters) ([]*Session, error)
}

type Event struct {
    ID          string
    SessionID   string
    WorkspaceID string
    Type        EventType
    Timestamp   time.Time
    URL         string
    Properties  map[string]interface{}
}
```

#### MetricsService
```go
type MetricsService interface {
    GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)
    GetKPIs(ctx context.Context, workspaceID string, timeRange TimeRange) (*KPIResponse, error)
    DetectAnomalies(ctx context.Context, workspaceID string) ([]*Anomaly, error)
}
```

### Enhanced Analysis Pipeline

#### Performance Analysis
```go
type PerformanceAnalyzer struct {
    client *http.Client
}

func (p *PerformanceAnalyzer) Analyze(url string, options AnalysisOptions) (*PerformanceMetrics, error) {
    // Measure load times, resource sizes, Core Web Vitals
    // Parse resource timing data
    // Calculate performance scores
}

type PerformanceMetrics struct {
    LoadTimeMS              int
    FirstContentfulPaint    int
    LargestContentfulPaint  int
    CumulativeLayoutShift   float64
    FirstInputDelay         int
    ResourceCount           int
    TotalSizeKB            int
    ImageOptimization      OptimizationScore
    CSSOptimization        OptimizationScore
    JSOptimization         OptimizationScore
}
```

#### SEO Analysis
```go
type SEOAnalyzer struct{}

func (s *SEOAnalyzer) Analyze(html []byte, headers http.Header) (*SEOMetrics, error) {
    // Parse HTML for meta tags, headings, structured data
    // Analyze content structure and optimization
    // Check for SEO best practices
}

type SEOMetrics struct {
    Title           string
    MetaDescription string
    H1Count         int
    H2Count         int
    StructuredData  []StructuredDataItem
    ImageAltTags    int
    InternalLinks   int
    ExternalLinks   int
    Score           int
    Issues          []SEOIssue
}
```

#### Accessibility Analysis
```go
type AccessibilityAnalyzer struct{}

func (a *AccessibilityAnalyzer) Analyze(html []byte) (*AccessibilityMetrics, error) {
    // Check WCAG compliance
    // Analyze color contrast, alt tags, ARIA labels
    // Identify accessibility issues
}

type AccessibilityMetrics struct {
    Score       int
    WCAGLevel   string
    Issues      []AccessibilityIssue
    ColorContrast struct {
        Passed int
        Failed int
    }
    AltTags struct {
        Present int
        Missing int
    }
}
```

#### Security Analysis
```go
type SecurityAnalyzer struct{}

func (s *SecurityAnalyzer) Analyze(headers http.Header, url string) (*SecurityMetrics, error) {
    // Check security headers
    // Analyze HTTPS configuration
    // Identify security vulnerabilities
}

type SecurityMetrics struct {
    HTTPS                bool
    HSTS                bool
    ContentSecurityPolicy bool
    XFrameOptions       bool
    XContentTypeOptions bool
    Score               int
    Issues              []SecurityIssue
}
```

### AI Insights Engine

#### Rules-Based Insight Generator
```go
type InsightGenerator struct {
    rules []InsightRule
}

type InsightRule interface {
    Evaluate(ctx context.Context, data *AnalysisData) (*Insight, error)
    Priority() Priority
    Type() InsightType
}

// Example rules
type PerformanceBottleneckRule struct{}
type ConversionFunnelRule struct{}
type SEOOptimizationRule struct{}
type AccessibilityIssueRule struct{}
```

## Data Models

### Database Schema

#### Analysis Results
```sql
CREATE TABLE analysis_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    session_id UUID,
    url TEXT NOT NULL,
    technologies JSONB,
    performance_metrics JSONB,
    seo_metrics JSONB,
    accessibility_metrics JSONB,
    security_metrics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_analysis_workspace_created ON analysis_results(workspace_id, created_at);
CREATE INDEX idx_analysis_session ON analysis_results(session_id);
CREATE INDEX idx_analysis_url ON analysis_results USING hash(url);
```

#### Sessions and Events
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    user_id TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    page_views INTEGER DEFAULT 0,
    events_count INTEGER DEFAULT 0,
    device_type TEXT,
    browser TEXT,
    country TEXT,
    referrer TEXT
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    workspace_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    url TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_workspace_timestamp ON events(workspace_id, timestamp);
CREATE INDEX idx_events_type ON events(event_type);
```

#### Insights
```sql
CREATE TABLE insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    insight_type TEXT NOT NULL,
    priority TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    impact_score INTEGER,
    effort_score INTEGER,
    recommendations JSONB,
    data_source JSONB,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_insights_workspace_status ON insights(workspace_id, status);
CREATE INDEX idx_insights_priority ON insights(priority);
```

#### Metrics Aggregations
```sql
CREATE TABLE daily_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    date DATE NOT NULL,
    total_sessions INTEGER DEFAULT 0,
    total_page_views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    bounce_rate DECIMAL(5,2),
    avg_session_duration INTEGER,
    conversion_rate DECIMAL(5,2),
    avg_load_time INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_daily_metrics_workspace_date ON daily_metrics(workspace_id, date);
```

## Error Handling

### Error Response Format
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Standard error codes
const (
    ErrCodeInvalidRequest   = "INVALID_REQUEST"
    ErrCodeUnauthorized     = "UNAUTHORIZED"
    ErrCodeRateLimited      = "RATE_LIMITED"
    ErrCodeInternalError    = "INTERNAL_ERROR"
    ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)
```

### Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    state       CircuitState
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    // Implement circuit breaker logic
    // Fail fast when service is down
    // Auto-recovery after timeout
}
```

## Testing Strategy

### Unit Testing
- Service layer unit tests with mocked dependencies
- Repository layer tests with test database
- Analysis engine tests with sample data
- Insight generation rule tests

### Integration Testing
- API endpoint tests with test database
- Database migration tests
- External service integration tests
- Performance benchmarking tests

### Load Testing
- Concurrent analysis request handling
- Database performance under load
- Memory usage and garbage collection
- Rate limiting effectiveness

### Test Data Management
```go
type TestDataManager struct {
    db *sql.DB
}

func (tdm *TestDataManager) SetupTestWorkspace() *Workspace {
    // Create test workspace with sample data
}

func (tdm *TestDataManager) CleanupTestData() {
    // Clean up test data after tests
}
```

## Performance Considerations

### Caching Strategy
- Redis for frequently accessed metrics
- In-memory caching for analysis rules
- HTTP response caching with appropriate TTL
- Database query result caching

### Database Optimization
- Proper indexing for time-series queries
- Partitioning for large tables
- Connection pooling and prepared statements
- Read replicas for analytics queries

### Concurrent Processing
- Worker pools for batch analysis
- Background job processing for insights
- Rate limiting per workspace
- Resource isolation between tenants

### Monitoring and Observability
```go
type Metrics struct {
    RequestDuration   prometheus.HistogramVec
    RequestCount      prometheus.CounterVec
    ActiveConnections prometheus.Gauge
    ErrorRate         prometheus.CounterVec
}

func (m *Metrics) RecordRequest(endpoint string, duration time.Duration, status int) {
    // Record metrics for monitoring
}
```