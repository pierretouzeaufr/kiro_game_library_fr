package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"board-game-library/internal/config"
	"board-game-library/internal/logging"
	"board-game-library/internal/routes"
	"board-game-library/pkg/database"
)

// App represents the main application
type App struct {
	config *config.Config
	logger *logging.Logger
	db     *database.DB
	server *http.Server
	router *gin.Engine
}

// New creates a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Set Gin mode based on log level
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	app := &App{
		config: cfg,
		logger: logger,
	}

	return app, nil
}

// Initialize initializes all application components
func (a *App) Initialize() error {
	a.logger.LogStartup(a.config)

	// Initialize database
	if err := a.initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize router
	if err := a.initializeRouter(); err != nil {
		return fmt.Errorf("failed to initialize router: %w", err)
	}

	// Initialize HTTP server
	a.initializeServer()

	return nil
}

// Run starts the application and blocks until shutdown
func (a *App) Run() error {
	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
		a.logger.LogServerStart(addr)
		serverErrors <- a.server.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		a.logger.Info("Received shutdown signal", "signal", sig.String())
		return a.Shutdown()
	}
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	a.logger.LogShutdown()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if a.server != nil {
		a.logger.LogServerStop()
		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Error("Failed to shutdown server gracefully", "error", err)
			return err
		}
	}

	// Close database connection
	if a.db != nil {
		a.logger.LogDatabaseDisconnection()
		if err := a.db.Close(); err != nil {
			a.logger.Error("Failed to close database connection", "error", err)
			return err
		}
	}

	a.logger.Info("Application shutdown completed")
	return nil
}

// initializeDatabase initializes the database connection
func (a *App) initializeDatabase() error {
	dbConfig := database.Config{
		DatabasePath: a.config.Database.Path,
	}

	db, err := database.Initialize(dbConfig)
	if err != nil {
		return err
	}

	// Configure connection pool
	db.SetMaxOpenConns(a.config.Database.MaxOpenConns)
	db.SetMaxIdleConns(a.config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(a.config.Database.ConnMaxLifetime)

	a.db = db
	a.logger.LogDatabaseConnection(a.config.Database.Path)
	return nil
}

// initializeRouter initializes the Gin router with all routes
func (a *App) initializeRouter() error {
	router := gin.New()

	// Add middleware
	router.Use(a.loggingMiddleware())
	router.Use(gin.Recovery())

	// Basic health check endpoint
	router.GET("/health", a.healthCheckHandler)

	// Status endpoint
	router.GET("/api/v1/status", a.statusHandler)

	// Setup all application routes
	if err := routes.SetupRoutes(router, a.db); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	a.router = router
	return nil
}

// initializeServer initializes the HTTP server
func (a *App) initializeServer() {
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      a.router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
	}
}

// loggingMiddleware creates a Gin middleware for request logging
func (a *App) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		a.logger.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// healthCheckHandler handles health check requests
func (a *App) healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "Board Game Library API is running",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// statusHandler handles status requests with more detailed information
func (a *App) statusHandler(c *gin.Context) {
	// Test database connection
	dbStatus := "ok"
	if err := a.db.Ping(); err != nil {
		dbStatus = "error"
		a.logger.Error("Database ping failed", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "Board Game Library API is running",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"database":  dbStatus,
		"config": gin.H{
			"server_port":     a.config.Server.Port,
			"database_path":   a.config.Database.Path,
			"alerts_enabled":  a.config.Alerts.EnableOverdue && a.config.Alerts.EnableReminders,
			"log_level":       a.config.Logging.Level,
		},
	})
}

// GetDB returns the database connection (for use by other components)
func (a *App) GetDB() *database.DB {
	return a.db
}

// GetLogger returns the logger (for use by other components)
func (a *App) GetLogger() *logging.Logger {
	return a.logger
}

// GetConfig returns the configuration (for use by other components)
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetRouter returns the Gin router (for use by serverless handlers)
func (a *App) GetRouter() *gin.Engine {
	return a.router
}