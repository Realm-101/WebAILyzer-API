-- Initial schema for WebAIlyzer Lite API
-- Migration: 001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Analysis Results Table
CREATE TABLE analysis_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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

-- Indexes for analysis_results
CREATE INDEX idx_analysis_workspace_created ON analysis_results(workspace_id, created_at);
CREATE INDEX idx_analysis_session ON analysis_results(session_id);
CREATE INDEX idx_analysis_url ON analysis_results USING hash(url);
CREATE INDEX idx_analysis_workspace_url ON analysis_results(workspace_id, url);

-- Sessions Table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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

-- Indexes for sessions
CREATE INDEX idx_sessions_workspace ON sessions(workspace_id);
CREATE INDEX idx_sessions_started_at ON sessions(started_at);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_workspace_started ON sessions(workspace_id, started_at);

-- Events Table
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    url TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for events
CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_workspace_timestamp ON events(workspace_id, timestamp);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_workspace_type ON events(workspace_id, event_type);

-- Insights Table
CREATE TABLE insights (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,
    insight_type TEXT NOT NULL,
    priority TEXT NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    title TEXT NOT NULL,
    description TEXT,
    impact_score INTEGER CHECK (impact_score >= 0 AND impact_score <= 100),
    effort_score INTEGER CHECK (effort_score >= 0 AND effort_score <= 100),
    recommendations JSONB,
    data_source JSONB,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'applied', 'dismissed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for insights
CREATE INDEX idx_insights_workspace_status ON insights(workspace_id, status);
CREATE INDEX idx_insights_priority ON insights(priority);
CREATE INDEX idx_insights_type ON insights(insight_type);
CREATE INDEX idx_insights_workspace_created ON insights(workspace_id, created_at);

-- Daily Metrics Table
CREATE TABLE daily_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,
    date DATE NOT NULL,
    total_sessions INTEGER DEFAULT 0,
    total_page_views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    bounce_rate DECIMAL(5,2) CHECK (bounce_rate >= 0 AND bounce_rate <= 100),
    avg_session_duration INTEGER,
    conversion_rate DECIMAL(5,2) CHECK (conversion_rate >= 0 AND conversion_rate <= 100),
    avg_load_time INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Unique constraint and indexes for daily_metrics
CREATE UNIQUE INDEX idx_daily_metrics_workspace_date ON daily_metrics(workspace_id, date);
CREATE INDEX idx_daily_metrics_date ON daily_metrics(date);

-- Add foreign key constraint for analysis_results session_id
ALTER TABLE analysis_results 
ADD CONSTRAINT fk_analysis_session 
FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL;

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_analysis_results_updated_at 
    BEFORE UPDATE ON analysis_results 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_insights_updated_at 
    BEFORE UPDATE ON insights 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();