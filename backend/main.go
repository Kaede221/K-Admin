package main

// @title K-Admin System API
// @version 1.0
// @description K-Admin 后台管理系统 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description JWT token format: Bearer {token}

import (
	"flag"
	"log"

	"k-admin-system/config"
	"k-admin-system/core"
	_ "k-admin-system/docs" // Swagger docs
	"k-admin-system/global"
	"k-admin-system/middleware"
	systemRouter "k-admin-system/router/system"
	toolsRouter "k-admin-system/router/tools"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Initialize Casbin enforcer
	casbinEnforcer, err := core.InitCasbin()
	if err != nil {
		logger.Fatal("Failed to initialize Casbin", zap.Error(err))
	}
	global.CasbinEnforcer = casbinEnforcer

	// Run database migrations
	if err := core.AutoMigrate(); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Initialize Gin router without default middleware
	r := gin.New()

	// Configure middleware chain in correct order
	// Order: Recovery → CORS → RateLimit → Logger → JWT → Casbin

	// 1. Recovery middleware (must be first to catch all panics)
	r.Use(middleware.Recovery())

	// 2. CORS middleware (handle cross-origin requests early)
	r.Use(middleware.CORS(cfg.CORS))

	// 3. Rate limiting middleware (prevent abuse before processing)
	r.Use(middleware.RateLimit(cfg.RateLimit))

	// 4. Logger middleware (log all requests)
	r.Use(middleware.Logger())

	// Health check endpoint (excluded from JWT and Casbin)
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
		systemRouter.InitRoleRouter(apiV1)
		systemRouter.InitMenuRouter(apiV1)

		// Tools module routes
		toolsGroup := apiV1.Group("/tools")
		toolsRouter.InitDBInspectorRouter(toolsGroup)
		toolsRouter.InitCodeGeneratorRouter(toolsGroup)
	}

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	logger.Info("Server starting", zap.String("port", cfg.Server.Port))
	if err := r.Run(cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
