package main

import (
	"flag"
	"log"

	"k-admin-system/config"
	"k-admin-system/core"
	"k-admin-system/global"
	systemRouter "k-admin-system/router/system"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to config file (YAML or JSON)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	global.Config = cfg

	// Initialize logger
	logger, err := core.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	global.Logger = logger
	defer core.SyncLogger(logger)

	logger.Info("Application starting",
		zap.String("mode", cfg.Server.Mode),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := core.InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	global.DB = db

	// Initialize Redis
	redisClient, err := core.InitRedis()
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	global.RedisClient = redisClient

	// Run database migrations
	if err := core.AutoMigrate(); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Initialize Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"mode":   cfg.Server.Mode,
		})
	})

	// API v1 routes
	apiV1 := r.Group("/api/v1")
	{
		// System module routes
		systemRouter.InitUserRouter(apiV1)
	}

	// Start server
	logger.Info("Server starting", zap.String("port", cfg.Server.Port))
	if err := r.Run(cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
