package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret            string `mapstructure:"secret"`
	AccessExpiration  int    `mapstructure:"access_expiration"`  // in minutes
	RefreshExpiration int    `mapstructure:"refresh_expiration"` // in days
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error, fatal
	Path       string `mapstructure:"path"`        // log file path
	MaxSize    int    `mapstructure:"max_size"`    // megabytes
	MaxAge     int    `mapstructure:"max_age"`     // days
	MaxBackups int    `mapstructure:"max_backups"` // number of backups
	Compress   bool   `mapstructure:"compress"`    // compress rotated files
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // in seconds
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool   `mapstructure:"enabled"`  // enable/disable rate limiting
	Requests int    `mapstructure:"requests"` // number of requests allowed
	Window   int    `mapstructure:"window"`   // time window in seconds
	KeyFunc  string `mapstructure:"key_func"` // "ip" or "user" - how to identify clients
}

// LoadConfig loads configuration from file and environment variables
// Supports YAML and JSON formats
// Environment variables take precedence over file configuration
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config file locations
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
	}

	// Enable reading from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("KADMIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional if all required values are in env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates that all required configuration fields are set
func validateConfig(config *Config) error {
	// Validate Server config
	if config.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug" // default mode
	}
	if config.Server.Mode != "debug" && config.Server.Mode != "release" && config.Server.Mode != "test" {
		return fmt.Errorf("server.mode must be one of: debug, release, test")
	}

	// Validate Database config
	if config.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if config.Database.Port == 0 {
		return fmt.Errorf("database.port is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if config.Database.Username == "" {
		return fmt.Errorf("database.username is required")
	}
	// Password can be empty for local development

	// Set default connection pool values if not specified
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 100
	}

	// Validate JWT config
	if config.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if config.JWT.AccessExpiration == 0 {
		config.JWT.AccessExpiration = 15 // default 15 minutes
	}
	if config.JWT.RefreshExpiration == 0 {
		config.JWT.RefreshExpiration = 7 // default 7 days
	}

	// Validate Redis config
	if config.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	if config.Redis.Port == 0 {
		return fmt.Errorf("redis.port is required")
	}
	// Password and DB can have default values

	// Validate Logger config
	if config.Logger.Level == "" {
		config.Logger.Level = "info" // default level
	}
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLevels[config.Logger.Level] {
		return fmt.Errorf("logger.level must be one of: debug, info, warn, error, fatal")
	}
	if config.Logger.Path == "" {
		config.Logger.Path = "./logs/app.log" // default path
	}
	// Set default log rotation values if not specified
	if config.Logger.MaxSize == 0 {
		config.Logger.MaxSize = 100 // 100MB
	}
	if config.Logger.MaxAge == 0 {
		config.Logger.MaxAge = 7 // 7 days
	}
	if config.Logger.MaxBackups == 0 {
		config.Logger.MaxBackups = 3
	}

	// Validate CORS config - set defaults if not specified
	if len(config.CORS.AllowOrigins) == 0 {
		config.CORS.AllowOrigins = []string{"*"} // default allow all origins
	}
	if len(config.CORS.AllowMethods) == 0 {
		config.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}
	if len(config.CORS.AllowHeaders) == 0 {
		config.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	}
	if config.CORS.MaxAge == 0 {
		config.CORS.MaxAge = 86400 // default 24 hours
	}

	// Validate RateLimit config - set defaults if not specified
	if config.RateLimit.Requests == 0 {
		config.RateLimit.Requests = 100 // default 100 requests
	}
	if config.RateLimit.Window == 0 {
		config.RateLimit.Window = 60 // default 60 seconds (1 minute)
	}
	if config.RateLimit.KeyFunc == "" {
		config.RateLimit.KeyFunc = "ip" // default to IP-based rate limiting
	}
	if config.RateLimit.KeyFunc != "ip" && config.RateLimit.KeyFunc != "user" {
		return fmt.Errorf("rate_limit.key_func must be one of: ip, user")
	}

	return nil
}
