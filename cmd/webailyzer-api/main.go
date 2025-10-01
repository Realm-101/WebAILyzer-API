package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	
	"github.com/projectdiscovery/wappalyzergo/internal/config"
	"github.com/projectdiscovery/wappalyzergo/internal/database"
	"github.com/projectdiscovery/wappalyzergo/internal/cache"
	"github.com/projectdiscovery/wappalyzergo/internal/handlers"
	"github.com/projectdiscovery/wappalyzergo/internal/middleware"
	"github.com/projectdiscovery/wappalyzergo/internal/monitoring"
	"github.com/projectdiscovery/wappalyzergo/internal/repositories/postgres"
	"github.com/projectdiscovery/wappalyzergo/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	// Initialize database connection
	dbConn, err := database.NewConnection(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbConn.Close()

	// Run database migrations
	if err := dbConn.Migrate(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}

	// Start database maintenance service
	maintenanceService := database.NewMaintenanceService(dbConn, logger)
	maintenanceCtx, cancelMaintenance := context.WithCancel(context.Background())
	maintenanceService.Start(maintenanceCtx)
	defer func() {
		cancelMaintenance()
		maintenanceService.Stop()
	}()

	// Initialize Redis cache
	var cacheService *cache.CacheService
	redisClient, err := cache.NewRedisClient(&cfg.Redis, logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
		cacheService = cache.NewCacheService(nil, logger) // Create service with nil client
	} else {
		cacheService = cache.NewCacheService(redisClient, logger)
		defer redisClient.Close()
	}

	// Initialize monitoring
	metricsCollector := monitoring.NewMetricsCollector()
	monitoringService := monitoring.NewMonitoringService(metricsCollector, logger, dbConn, cacheService)
	
	// Start monitoring service
	monitoringCtx, cancelMonitoring := context.WithCancel(context.Background())
	monitoringService.Start(monitoringCtx)
	defer func() {
		cancelMonitoring()
		monitoringService.Stop()
	}()

	// Setup HTTP server
	server, rateLimitMiddleware := setupServer(cfg, logger, dbConn, cacheService, metricsCollector)
	defer rateLimitMiddleware.Stop()

	// Start server
	go func() {
		logger.WithField("address", cfg.Server.GetServerAddr()).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown complete")
	}
}

func setupLogger(cfg config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	return logger
}

func setupServer(cfg *config.Config, logger *logrus.Logger, dbConn *database.Connection, cacheService *cache.CacheService, metricsCollector *monitoring.MetricsCollector) (*http.Server, *middleware.RateLimitMiddleware) {
	router := mux.NewRouter()

	// Setup middleware
	metricsMiddleware := middleware.NewMetricsMiddleware(metricsCollector, logger)
	router.Use(metricsMiddleware.CollectMetrics)
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())

	// Initialize repositories
	analysisRepo := postgres.NewAnalysisRepository(dbConn.Pool)
	metricsRepo := postgres.NewMetricsRepository(dbConn.Pool)
	sessionRepo := postgres.NewSessionRepository(dbConn.Pool)
	eventRepo := postgres.NewEventRepository(dbConn.Pool)
	insightRepo := postgres.NewInsightRepository(dbConn.Pool)
	workspaceRepo := postgres.NewWorkspaceRepository(dbConn.Pool)
	
	// Initialize services
	analysisService, err := services.NewAnalysisService(analysisRepo, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize analysis service")
	}

	metricsService := services.NewMetricsService(metricsRepo, sessionRepo, eventRepo, analysisRepo)
	insightsService := services.NewInsightsService(insightRepo)
	exportService := services.NewExportService(analysisRepo, metricsRepo, sessionRepo, eventRepo, logger)

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(workspaceRepo, logger)
	
	// Initialize rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(&cfg.RateLimit, logger)

	// Setup handlers
	healthHandler := handlers.NewHealthHandler(logger, dbConn, cacheService)
	healthHandler.RegisterRoutes(router)
	
	// Add Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Create API subrouter with authentication and rate limiting
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.Authenticate)
	apiRouter.Use(rateLimitMiddleware.RateLimit)

	analysisHandler := handlers.NewAnalysisHandler(analysisService, cacheService, logger)
	analysisHandler.RegisterRoutes(apiRouter)

	metricsHandler := handlers.NewMetricsHandler(metricsService, cacheService, logger)
	metricsHandler.RegisterRoutes(apiRouter)

	insightsHandler := handlers.NewInsightsHandler(insightsService, logger)
	insightsHandler.RegisterRoutes(apiRouter)

	exportHandler := handlers.NewExportHandler(exportService, logger)
	exportHandler.RegisterRoutes(apiRouter)

	// Setup CORS
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"*"}),
	)(router)

	server := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      corsHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return server, rateLimitMiddleware
}

func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Call the next handler
			next.ServeHTTP(w, r)
			
			// Log the request
			logger.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start),
				"remote":   r.RemoteAddr,
			}).Info("HTTP request")
		})
	}
}

func corsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}