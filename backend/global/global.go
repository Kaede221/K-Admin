package global

import (
	"k-admin-system/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Global variables accessible throughout the application
var (
	// Config holds the application configuration
	Config *config.Config

	// Logger holds the global Zap logger instance
	Logger *zap.Logger

	// DB holds the global Gorm database instance
	DB *gorm.DB

	// RedisClient holds the global Redis client instance
	RedisClient *redis.Client
)
