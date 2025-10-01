-- Performance optimizations for WebAIlyzer Lite API
-- Migration: 002_performance_optimizations.sql

-- Add additional indexes for time-series queries and common patterns

-- Composite indexes for analysis_results time-series queries
CREATE INDEX idx_analysis_workspace_created_desc ON analysis_results(workspace_id, created_at DESC);
CREATE INDEX idx_analysis_session_created ON analysis_results(session_id, created_at) WHERE session_id IS NOT NULL;

-- JSONB indexes for technology detection queries
CREATE INDEX idx_analysis_technologies_gin ON analysis_results USING gin(technologies);
CREATE INDEX idx_analysis_performance_gin ON analysis_results USING gin(performance_metrics);
CREATE INDEX idx_analysis_seo_gin ON analysis_results USING gin(seo_metrics);

-- Partial indexes for active sessions
CREATE INDEX idx_sessions_active ON sessions(workspace_id, started_at) WHERE ended_at IS NULL;
CREATE INDEX idx_sessions_recent ON sessions(workspace_id, started_at DESC) WHERE started_at > NOW() - INTERVAL '30 days';

-- Events time-series optimizations
CREATE INDEX idx_events_workspace_timestamp_desc ON events(workspace_id, timestamp DESC);
CREATE INDEX idx_events_session_timestamp ON events(session_id, timestamp);
CREATE INDEX idx_events_recent ON events(workspace_id, event_type, timestamp) WHERE timestamp > NOW() - INTERVAL '7 days';

-- JSONB index for event properties
CREATE INDEX idx_events_properties_gin ON events USING gin(properties);

-- Insights optimization indexes
CREATE INDEX idx_insights_workspace_priority_created ON insights(workspace_id, priority, created_at DESC);
CREATE INDEX idx_insights_status_created ON insights(status, created_at DESC) WHERE status = 'pending';

-- Daily metrics time-series indexes
CREATE INDEX idx_daily_metrics_workspace_date_desc ON daily_metrics(workspace_id, date DESC);
CREATE INDEX idx_daily_metrics_date_desc ON daily_metrics(date DESC);

-- Create materialized view for workspace statistics
CREATE MATERIALIZED VIEW workspace_stats AS
SELECT 
    workspace_id,
    COUNT(DISTINCT id) as total_analyses,
    COUNT(DISTINCT session_id) as total_sessions,
    COUNT(DISTINCT DATE(created_at)) as active_days,
    MIN(created_at) as first_analysis,
    MAX(created_at) as last_analysis,
    AVG(CASE WHEN performance_metrics->>'load_time_ms' IS NOT NULL 
        THEN (performance_metrics->>'load_time_ms')::numeric 
        ELSE NULL END) as avg_load_time,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as analyses_last_7_days,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as analyses_last_30_days
FROM analysis_results 
GROUP BY workspace_id;

-- Index for the materialized view
CREATE UNIQUE INDEX idx_workspace_stats_workspace_id ON workspace_stats(workspace_id);

-- Create function to refresh workspace stats
CREATE OR REPLACE FUNCTION refresh_workspace_stats()
RETURNS void AS $
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_stats;
END;
$ language 'plpgsql';

-- Create hourly metrics aggregation table for better performance
CREATE TABLE hourly_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,
    hour_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    total_sessions INTEGER DEFAULT 0,
    total_page_views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    total_events INTEGER DEFAULT 0,
    avg_load_time INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for hourly metrics
CREATE UNIQUE INDEX idx_hourly_metrics_workspace_hour ON hourly_metrics(workspace_id, hour_timestamp);
CREATE INDEX idx_hourly_metrics_hour_desc ON hourly_metrics(hour_timestamp DESC);

-- Create function to aggregate hourly metrics
CREATE OR REPLACE FUNCTION aggregate_hourly_metrics(target_hour TIMESTAMP WITH TIME ZONE)
RETURNS void AS $
DECLARE
    start_time TIMESTAMP WITH TIME ZONE;
    end_time TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Calculate hour boundaries
    start_time := date_trunc('hour', target_hour);
    end_time := start_time + INTERVAL '1 hour';
    
    -- Insert or update hourly metrics
    INSERT INTO hourly_metrics (workspace_id, hour_timestamp, total_sessions, total_page_views, unique_visitors, total_events, avg_load_time)
    SELECT 
        s.workspace_id,
        start_time,
        COUNT(DISTINCT s.id) as total_sessions,
        COALESCE(SUM(s.page_views), 0) as total_page_views,
        COUNT(DISTINCT s.user_id) FILTER (WHERE s.user_id IS NOT NULL) as unique_visitors,
        COUNT(e.id) as total_events,
        AVG((ar.performance_metrics->>'load_time_ms')::numeric) FILTER (WHERE ar.performance_metrics->>'load_time_ms' IS NOT NULL) as avg_load_time
    FROM sessions s
    LEFT JOIN events e ON s.id = e.session_id AND e.timestamp >= start_time AND e.timestamp < end_time
    LEFT JOIN analysis_results ar ON s.id = ar.session_id AND ar.created_at >= start_time AND ar.created_at < end_time
    WHERE s.started_at >= start_time AND s.started_at < end_time
    GROUP BY s.workspace_id
    ON CONFLICT (workspace_id, hour_timestamp) 
    DO UPDATE SET
        total_sessions = EXCLUDED.total_sessions,
        total_page_views = EXCLUDED.total_page_views,
        unique_visitors = EXCLUDED.unique_visitors,
        total_events = EXCLUDED.total_events,
        avg_load_time = EXCLUDED.avg_load_time,
        created_at = NOW();
END;
$ language 'plpgsql';

-- Create indexes for URL pattern analysis
CREATE INDEX idx_analysis_url_pattern ON analysis_results(workspace_id, substring(url from '^https?://[^/]+')) WHERE url IS NOT NULL;

-- Create partial index for failed analyses (if we track them)
CREATE INDEX idx_analysis_errors ON analysis_results(workspace_id, created_at) WHERE technologies IS NULL;

-- Add database statistics collection
CREATE OR REPLACE FUNCTION collect_table_stats()
RETURNS TABLE(
    table_name text,
    row_count bigint,
    table_size text,
    index_size text,
    total_size text
) AS $
BEGIN
    RETURN QUERY
    SELECT 
        schemaname||'.'||tablename as table_name,
        n_tup_ins - n_tup_del as row_count,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as table_size,
        pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as index_size,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) + pg_indexes_size(schemaname||'.'||tablename)) as total_size
    FROM pg_stat_user_tables 
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
END;
$ language 'plpgsql';

-- Create function for query performance analysis
CREATE OR REPLACE FUNCTION get_slow_queries()
RETURNS TABLE(
    query text,
    calls bigint,
    total_time double precision,
    mean_time double precision,
    rows bigint
) AS $
BEGIN
    RETURN QUERY
    SELECT 
        pg_stat_statements.query,
        pg_stat_statements.calls,
        pg_stat_statements.total_exec_time,
        pg_stat_statements.mean_exec_time,
        pg_stat_statements.rows
    FROM pg_stat_statements 
    WHERE pg_stat_statements.query NOT LIKE '%pg_stat_statements%'
    ORDER BY pg_stat_statements.mean_exec_time DESC
    LIMIT 20;
EXCEPTION
    WHEN undefined_table THEN
        RAISE NOTICE 'pg_stat_statements extension not available';
        RETURN;
END;
$ language 'plpgsql';