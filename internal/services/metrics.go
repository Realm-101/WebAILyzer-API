package services

import (
	"context"
	"fmt"
	"time"
	"github.com/projectdiscovery/wappalyzergo/internal/models"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories"
	"github.com/google/uuid"
)

// metricsService implements the MetricsService interface
type metricsService struct {
	metricsRepo repositories.MetricsRepository
	sessionRepo repositories.SessionRepository
	eventRepo   repositories.EventRepository
	analysisRepo repositories.AnalysisRepository
}

// NewMetricsService creates a new metrics service
func NewMetricsService(
	metricsRepo repositories.MetricsRepository,
	sessionRepo repositories.SessionRepository,
	eventRepo repositories.EventRepository,
	analysisRepo repositories.AnalysisRepository,
) MetricsService {
	return &metricsService{
		metricsRepo:  metricsRepo,
		sessionRepo:  sessionRepo,
		eventRepo:    eventRepo,
		analysisRepo: analysisRepo,
	}
}

// GetMetrics retrieves metrics for a workspace within the specified time range and granularity
func (s *metricsService) GetMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	if err := s.validateMetricsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid metrics request: %w", err)
	}

	switch req.Granularity {
	case "daily":
		return s.getDailyMetrics(ctx, req)
	case "hourly":
		return s.getHourlyMetrics(ctx, req)
	case "weekly":
		return s.getWeeklyMetrics(ctx, req)
	case "monthly":
		return s.getMonthlyMetrics(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported granularity: %s", req.Granularity)
	}
}

// GetKPIs retrieves key performance indicators for a workspace
func (s *metricsService) GetKPIs(ctx context.Context, workspaceID uuid.UUID, timeRange TimeRange) (*KPIResponse, error) {
	// Calculate current period metrics
	currentMetrics, err := s.calculateRealTimeMetrics(ctx, workspaceID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate current metrics: %w", err)
	}

	// Calculate previous period for comparison
	duration := timeRange.End.Sub(timeRange.Start)
	prevStart := timeRange.Start.Add(-duration)
	prevEnd := timeRange.Start

	_, err = s.calculateRealTimeMetrics(ctx, workspaceID, prevStart, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate previous metrics: %w", err)
	}

	kpis := []models.KPI{
		{
			Name:        "Conversion Rate",
			Value:       currentMetrics.ConversionRate,
			Target:      5.0, // 5% target conversion rate
			Status:      s.getKPIStatus(currentMetrics.ConversionRate, 5.0),
			Description: "Percentage of sessions that result in conversions",
		},
		{
			Name:        "Bounce Rate",
			Value:       currentMetrics.BounceRate,
			Target:      40.0, // 40% target bounce rate (lower is better)
			Status:      s.getKPIStatus(40.0, currentMetrics.BounceRate), // Inverted for bounce rate
			Description: "Percentage of single-page sessions",
		},
		{
			Name:        "Average Session Duration",
			Value:       float64(currentMetrics.AvgSessionDuration),
			Target:      180.0, // 3 minutes target
			Status:      s.getKPIStatus(float64(currentMetrics.AvgSessionDuration), 180.0),
			Description: "Average time users spend on the site (seconds)",
		},
		{
			Name:        "Page Load Time",
			Value:       float64(currentMetrics.AvgLoadTime),
			Target:      2000.0, // 2 seconds target (lower is better)
			Status:      s.getKPIStatus(2000.0, float64(currentMetrics.AvgLoadTime)), // Inverted for load time
			Description: "Average page load time (milliseconds)",
		},
	}

	return &KPIResponse{
		KPIs:      kpis,
		Timestamp: time.Now(),
	}, nil
}

// DetectAnomalies detects anomalies in metrics data
func (s *metricsService) DetectAnomalies(ctx context.Context, workspaceID uuid.UUID) ([]*models.Anomaly, error) {
	// Get last 30 days of daily metrics for baseline
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	dailyMetrics, err := s.metricsRepo.GetDailyMetrics(ctx, workspaceID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	if len(dailyMetrics) < 7 {
		// Not enough data for anomaly detection
		return []*models.Anomaly{}, nil
	}

	var anomalies []*models.Anomaly

	// Detect anomalies in conversion rate
	conversionAnomalies := s.detectMetricAnomalies(dailyMetrics, "conversion_rate", func(m *models.DailyMetrics) float64 {
		if m.ConversionRate != nil {
			return *m.ConversionRate
		}
		return 0
	})
	anomalies = append(anomalies, conversionAnomalies...)

	// Detect anomalies in bounce rate
	bounceAnomalies := s.detectMetricAnomalies(dailyMetrics, "bounce_rate", func(m *models.DailyMetrics) float64 {
		if m.BounceRate != nil {
			return *m.BounceRate
		}
		return 0
	})
	anomalies = append(anomalies, bounceAnomalies...)

	// Detect anomalies in session duration
	durationAnomalies := s.detectMetricAnomalies(dailyMetrics, "avg_session_duration", func(m *models.DailyMetrics) float64 {
		if m.AvgSessionDuration != nil {
			return float64(*m.AvgSessionDuration)
		}
		return 0
	})
	anomalies = append(anomalies, durationAnomalies...)

	return anomalies, nil
}

// CalculateAndStoreDailyMetrics calculates and stores daily metrics for a workspace
func (s *metricsService) CalculateAndStoreDailyMetrics(ctx context.Context, workspaceID uuid.UUID, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	metrics, err := s.calculateRealTimeMetrics(ctx, workspaceID, startOfDay, endOfDay)
	if err != nil {
		return fmt.Errorf("failed to calculate daily metrics: %w", err)
	}

	dailyMetrics := &models.DailyMetrics{
		ID:                   uuid.New(),
		WorkspaceID:          workspaceID,
		Date:                 startOfDay,
		TotalSessions:        metrics.TotalSessions,
		TotalPageViews:       metrics.TotalPageViews,
		UniqueVisitors:       metrics.UniqueVisitors,
		BounceRate:           &metrics.BounceRate,
		AvgSessionDuration:   &metrics.AvgSessionDuration,
		ConversionRate:       &metrics.ConversionRate,
		AvgLoadTime:          &metrics.AvgLoadTime,
		CreatedAt:            time.Now(),
	}

	return s.metricsRepo.CreateDailyMetrics(ctx, dailyMetrics)
}

// validateMetricsRequest validates the metrics request
func (s *metricsService) validateMetricsRequest(req *models.MetricsRequest) error {
	if req.WorkspaceID == uuid.Nil {
		return fmt.Errorf("workspace_id is required")
	}
	if req.StartDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}
	if req.EndDate.IsZero() {
		return fmt.Errorf("end_date is required")
	}
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}
	if req.Granularity == "" {
		return fmt.Errorf("granularity is required")
	}
	return nil
}

// getDailyMetrics retrieves daily aggregated metrics
func (s *metricsService) getDailyMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	dailyMetrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	// Calculate previous period for comparison
	duration := req.EndDate.Sub(req.StartDate)
	prevStart := req.StartDate.Add(-duration)
	prevEnd := req.StartDate

	previousMetrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, prevStart, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous metrics: %w", err)
	}

	return s.buildMetricsResponse(dailyMetrics, previousMetrics), nil
}

// getHourlyMetrics calculates hourly metrics from real-time data
func (s *metricsService) getHourlyMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	var dataPoints []models.DataPoint
	current := req.StartDate

	for current.Before(req.EndDate) {
		hourEnd := current.Add(time.Hour)
		if hourEnd.After(req.EndDate) {
			hourEnd = req.EndDate
		}

		metrics, err := s.calculateRealTimeMetrics(ctx, req.WorkspaceID, current, hourEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hourly metrics: %w", err)
		}

		dataPoints = append(dataPoints, models.DataPoint{
			Timestamp: current,
			Value:     metrics.ConversionRate,
		})

		current = hourEnd
	}

	// Build response with hourly data points
	response := &models.MetricsResponse{
		Metrics: map[string]models.MetricData{
			"conversion_rate": {
				DataPoints: dataPoints,
			},
		},
		KPIs:      []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	return response, nil
}

// getWeeklyMetrics aggregates daily metrics into weekly periods
func (s *metricsService) getWeeklyMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	dailyMetrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	weeklyData := s.aggregateToWeekly(dailyMetrics)
	
	response := &models.MetricsResponse{
		Metrics: map[string]models.MetricData{
			"conversion_rate": {
				DataPoints: weeklyData,
			},
		},
		KPIs:      []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	return response, nil
}

// getMonthlyMetrics aggregates daily metrics into monthly periods
func (s *metricsService) getMonthlyMetrics(ctx context.Context, req *models.MetricsRequest) (*models.MetricsResponse, error) {
	dailyMetrics, err := s.metricsRepo.GetDailyMetrics(ctx, req.WorkspaceID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	monthlyData := s.aggregateToMonthly(dailyMetrics)
	
	response := &models.MetricsResponse{
		Metrics: map[string]models.MetricData{
			"conversion_rate": {
				DataPoints: monthlyData,
			},
		},
		KPIs:      []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	return response, nil
}

// calculateRealTimeMetrics calculates metrics from raw session and event data
func (s *metricsService) calculateRealTimeMetrics(ctx context.Context, workspaceID uuid.UUID, startTime, endTime time.Time) (*realTimeMetrics, error) {
	// Get sessions for the time period
	sessionFilters := &SessionFilters{
		WorkspaceID: workspaceID,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Limit:       10000, // Large limit to get all sessions
	}

	sessions, err := s.sessionRepo.GetSessionsByWorkspace(ctx, workspaceID, sessionFilters.Limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Filter sessions by time range
	var filteredSessions []*models.Session
	for _, session := range sessions {
		if session.StartedAt.After(startTime) && session.StartedAt.Before(endTime) {
			filteredSessions = append(filteredSessions, session)
		}
	}

	// Get events for conversion calculation
	events, err := s.eventRepo.GetEventsByWorkspace(ctx, workspaceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Get analysis results for load time calculation
	analysisFilters := &repositories.AnalysisFilters{
		WorkspaceID: workspaceID,
		StartDate:   &startTime,
		EndDate:     &endTime,
		Limit:       10000,
	}

	analysisResults, err := s.analysisRepo.GetByFilters(ctx, analysisFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis results: %w", err)
	}

	return s.calculateMetricsFromData(filteredSessions, events, analysisResults), nil
}

// calculateMetricsFromData calculates metrics from raw data
func (s *metricsService) calculateMetricsFromData(sessions []*models.Session, events []*models.Event, analyses []*models.AnalysisResult) *realTimeMetrics {
	metrics := &realTimeMetrics{}

	if len(sessions) == 0 {
		return metrics
	}

	// Basic session metrics
	metrics.TotalSessions = len(sessions)
	
	// Calculate unique visitors (simplified - using user_id if available)
	uniqueUsers := make(map[string]bool)
	totalDuration := 0
	bouncedSessions := 0
	totalPageViews := 0

	for _, session := range sessions {
		if session.UserID != nil && *session.UserID != "" {
			uniqueUsers[*session.UserID] = true
		}
		
		if session.DurationSeconds != nil {
			totalDuration += *session.DurationSeconds
		}
		
		totalPageViews += session.PageViews
		if session.PageViews <= 1 {
			bouncedSessions++
		}
	}

	metrics.UniqueVisitors = len(uniqueUsers)
	if metrics.UniqueVisitors == 0 {
		metrics.UniqueVisitors = metrics.TotalSessions // Fallback if no user IDs
	}

	metrics.TotalPageViews = totalPageViews
	
	if metrics.TotalSessions > 0 {
		metrics.AvgSessionDuration = totalDuration / metrics.TotalSessions
		metrics.BounceRate = (float64(bouncedSessions) / float64(metrics.TotalSessions)) * 100
	}

	// Calculate conversion rate from events
	conversionEvents := 0
	for _, event := range events {
		if event.EventType == "conversion" {
			conversionEvents++
		}
	}

	if metrics.TotalSessions > 0 {
		metrics.ConversionRate = (float64(conversionEvents) / float64(metrics.TotalSessions)) * 100
	}

	// Calculate average load time from analysis results
	if len(analyses) > 0 {
		totalLoadTime := 0
		validLoadTimes := 0

		for _, analysis := range analyses {
			if analysis.PerformanceMetrics != nil {
				if loadTime, ok := analysis.PerformanceMetrics["load_time_ms"]; ok {
					if loadTimeFloat, ok := loadTime.(float64); ok {
						totalLoadTime += int(loadTimeFloat)
						validLoadTimes++
					}
				}
			}
		}

		if validLoadTimes > 0 {
			metrics.AvgLoadTime = totalLoadTime / validLoadTimes
		}
	}

	return metrics
}

// buildMetricsResponse builds a metrics response from daily metrics data
func (s *metricsService) buildMetricsResponse(current, previous []*models.DailyMetrics) *models.MetricsResponse {
	response := &models.MetricsResponse{
		Metrics:   make(map[string]models.MetricData),
		KPIs:      []models.KPI{},
		Anomalies: []models.Anomaly{},
	}

	// Calculate aggregated values for current and previous periods
	currentAgg := s.aggregateDailyMetrics(current)
	previousAgg := s.aggregateDailyMetrics(previous)

	// Build data points for current period
	var conversionDataPoints []models.DataPoint
	var bounceDataPoints []models.DataPoint
	var durationDataPoints []models.DataPoint
	var loadTimeDataPoints []models.DataPoint

	for _, metric := range current {
		timestamp := metric.Date
		
		if metric.ConversionRate != nil {
			conversionDataPoints = append(conversionDataPoints, models.DataPoint{
				Timestamp: timestamp,
				Value:     *metric.ConversionRate,
			})
		}
		
		if metric.BounceRate != nil {
			bounceDataPoints = append(bounceDataPoints, models.DataPoint{
				Timestamp: timestamp,
				Value:     *metric.BounceRate,
			})
		}
		
		if metric.AvgSessionDuration != nil {
			durationDataPoints = append(durationDataPoints, models.DataPoint{
				Timestamp: timestamp,
				Value:     float64(*metric.AvgSessionDuration),
			})
		}
		
		if metric.AvgLoadTime != nil {
			loadTimeDataPoints = append(loadTimeDataPoints, models.DataPoint{
				Timestamp: timestamp,
				Value:     float64(*metric.AvgLoadTime),
			})
		}
	}

	// Build metrics with trend information
	response.Metrics["conversion_rate"] = models.MetricData{
		Current:    currentAgg.ConversionRate,
		Previous:   previousAgg.ConversionRate,
		Trend:      s.calculateTrend(currentAgg.ConversionRate, previousAgg.ConversionRate),
		DataPoints: conversionDataPoints,
	}

	response.Metrics["bounce_rate"] = models.MetricData{
		Current:    currentAgg.BounceRate,
		Previous:   previousAgg.BounceRate,
		Trend:      s.calculateTrend(previousAgg.BounceRate, currentAgg.BounceRate), // Inverted for bounce rate
		DataPoints: bounceDataPoints,
	}

	response.Metrics["avg_session_duration"] = models.MetricData{
		Current:    float64(currentAgg.AvgSessionDuration),
		Previous:   float64(previousAgg.AvgSessionDuration),
		Trend:      s.calculateTrend(float64(currentAgg.AvgSessionDuration), float64(previousAgg.AvgSessionDuration)),
		DataPoints: durationDataPoints,
	}

	response.Metrics["avg_load_time"] = models.MetricData{
		Current:    float64(currentAgg.AvgLoadTime),
		Previous:   float64(previousAgg.AvgLoadTime),
		Trend:      s.calculateTrend(float64(previousAgg.AvgLoadTime), float64(currentAgg.AvgLoadTime)), // Inverted for load time
		DataPoints: loadTimeDataPoints,
	}

	return response
}

// aggregateDailyMetrics aggregates a slice of daily metrics into a single summary
func (s *metricsService) aggregateDailyMetrics(metrics []*models.DailyMetrics) *aggregatedMetrics {
	if len(metrics) == 0 {
		return &aggregatedMetrics{}
	}

	agg := &aggregatedMetrics{}
	
	totalSessions := 0
	totalPageViews := 0
	totalUniqueVisitors := 0
	
	var bounceRates []float64
	var sessionDurations []int
	var conversionRates []float64
	var loadTimes []int

	for _, metric := range metrics {
		totalSessions += metric.TotalSessions
		totalPageViews += metric.TotalPageViews
		totalUniqueVisitors += metric.UniqueVisitors

		if metric.BounceRate != nil {
			bounceRates = append(bounceRates, *metric.BounceRate)
		}
		if metric.AvgSessionDuration != nil {
			sessionDurations = append(sessionDurations, *metric.AvgSessionDuration)
		}
		if metric.ConversionRate != nil {
			conversionRates = append(conversionRates, *metric.ConversionRate)
		}
		if metric.AvgLoadTime != nil {
			loadTimes = append(loadTimes, *metric.AvgLoadTime)
		}
	}

	agg.TotalSessions = totalSessions
	agg.TotalPageViews = totalPageViews
	agg.UniqueVisitors = totalUniqueVisitors

	// Calculate averages
	if len(bounceRates) > 0 {
		sum := 0.0
		for _, rate := range bounceRates {
			sum += rate
		}
		agg.BounceRate = sum / float64(len(bounceRates))
	}

	if len(sessionDurations) > 0 {
		sum := 0
		for _, duration := range sessionDurations {
			sum += duration
		}
		agg.AvgSessionDuration = sum / len(sessionDurations)
	}

	if len(conversionRates) > 0 {
		sum := 0.0
		for _, rate := range conversionRates {
			sum += rate
		}
		agg.ConversionRate = sum / float64(len(conversionRates))
	}

	if len(loadTimes) > 0 {
		sum := 0
		for _, time := range loadTimes {
			sum += time
		}
		agg.AvgLoadTime = sum / len(loadTimes)
	}

	return agg
}

// aggregateToWeekly aggregates daily metrics into weekly data points
func (s *metricsService) aggregateToWeekly(dailyMetrics []*models.DailyMetrics) []models.DataPoint {
	if len(dailyMetrics) == 0 {
		return []models.DataPoint{}
	}

	weeklyData := make(map[string][]*models.DailyMetrics)
	
	// Group by week
	for _, metric := range dailyMetrics {
		year, week := metric.Date.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%d", year, week)
		weeklyData[weekKey] = append(weeklyData[weekKey], metric)
	}

	var dataPoints []models.DataPoint
	for _, weekMetrics := range weeklyData {
		agg := s.aggregateDailyMetrics(weekMetrics)
		
		// Use the first day of the week as timestamp
		firstDay := weekMetrics[0].Date
		for _, metric := range weekMetrics {
			if metric.Date.Before(firstDay) {
				firstDay = metric.Date
			}
		}

		dataPoints = append(dataPoints, models.DataPoint{
			Timestamp: firstDay,
			Value:     agg.ConversionRate,
		})
	}

	return dataPoints
}

// aggregateToMonthly aggregates daily metrics into monthly data points
func (s *metricsService) aggregateToMonthly(dailyMetrics []*models.DailyMetrics) []models.DataPoint {
	if len(dailyMetrics) == 0 {
		return []models.DataPoint{}
	}

	monthlyData := make(map[string][]*models.DailyMetrics)
	
	// Group by month
	for _, metric := range dailyMetrics {
		monthKey := metric.Date.Format("2006-01")
		monthlyData[monthKey] = append(monthlyData[monthKey], metric)
	}

	var dataPoints []models.DataPoint
	for _, monthMetrics := range monthlyData {
		agg := s.aggregateDailyMetrics(monthMetrics)
		
		// Use the first day of the month as timestamp
		firstDay := monthMetrics[0].Date
		for _, metric := range monthMetrics {
			if metric.Date.Before(firstDay) {
				firstDay = metric.Date
			}
		}

		dataPoints = append(dataPoints, models.DataPoint{
			Timestamp: firstDay,
			Value:     agg.ConversionRate,
		})
	}

	return dataPoints
}

// detectMetricAnomalies detects anomalies in a specific metric using simple statistical methods
func (s *metricsService) detectMetricAnomalies(metrics []*models.DailyMetrics, metricName string, valueExtractor func(*models.DailyMetrics) float64) []*models.Anomaly {
	if len(metrics) < 7 {
		return []*models.Anomaly{}
	}

	values := make([]float64, len(metrics))
	for i, metric := range metrics {
		values[i] = valueExtractor(metric)
	}

	// Calculate mean and standard deviation
	mean := s.calculateMean(values)
	stdDev := s.calculateStdDev(values, mean)

	var anomalies []*models.Anomaly
	threshold := 2.0 // 2 standard deviations

	// Check last few days for anomalies
	for i := len(metrics) - 3; i < len(metrics); i++ {
		if i < 0 {
			continue
		}

		value := values[i]
		deviation := (value - mean) / stdDev

		if deviation > threshold || deviation < -threshold {
			severity := "medium"
			if deviation > 3.0 || deviation < -3.0 {
				severity = "high"
			}

			anomalies = append(anomalies, &models.Anomaly{
				Metric:      metricName,
				Timestamp:   metrics[i].Date,
				Expected:    mean,
				Actual:      value,
				Severity:    severity,
				Description: fmt.Sprintf("%s is %.1f standard deviations from the mean", metricName, deviation),
			})
		}
	}

	return anomalies
}

// calculateMean calculates the mean of a slice of float64 values
func (s *metricsService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation of a slice of float64 values
func (s *metricsService) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquaredDiffs := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return variance // Using variance instead of sqrt for simplicity
}

// calculateTrend determines if a metric is trending up, down, or stable
func (s *metricsService) calculateTrend(current, previous float64) string {
	if previous == 0 {
		if current > 0 {
			return "up"
		}
		return "stable"
	}

	change := ((current - previous) / previous) * 100
	if change > 5 {
		return "up"
	} else if change < -5 {
		return "down"
	}
	return "stable"
}

// getKPIStatus determines the status of a KPI based on its value and target
func (s *metricsService) getKPIStatus(value, target float64) string {
	ratio := value / target
	if ratio >= 0.9 {
		return "good"
	} else if ratio >= 0.7 {
		return "warning"
	}
	return "critical"
}

// realTimeMetrics represents calculated metrics from raw data
type realTimeMetrics struct {
	TotalSessions      int
	TotalPageViews     int
	UniqueVisitors     int
	BounceRate         float64
	AvgSessionDuration int
	ConversionRate     float64
	AvgLoadTime        int
}

// aggregatedMetrics represents aggregated metrics data
type aggregatedMetrics struct {
	TotalSessions      int
	TotalPageViews     int
	UniqueVisitors     int
	BounceRate         float64
	AvgSessionDuration int
	ConversionRate     float64
	AvgLoadTime        int
}