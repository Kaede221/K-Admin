package config

import (
	"os"
	"testing"
)

func TestLoadConfig_YAML(t *testing.T) {
	// Create a temporary YAML config file
	content := `
server:
  port: ":9090"
  mode: "release"
database:
  host: "testdb"
  port: 3307
  name: "test_db"
  username: "testuser"
  password: "testpass"
jwt:
  secret: "test-secret"
  access_expiration: 30
  refresh_expiration: 14
redis:
  host: "testredis"
  port: 6380
  password: "redispass"
  db: 1
logger:
  level: "debug"
  path: "./test.log"
  max_size: 50
  max_age: 3
  max_backups: 2
  compress: false
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify Server config
	if cfg.Server.Port != ":9090" {
		t.Errorf("Expected port :9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Expected mode release, got %s", cfg.Server.Mode)
	}

	// Verify Database config
	if cfg.Database.Host != "testdb" {
		t.Errorf("Expected host testdb, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 3307 {
		t.Errorf("Expected port 3307, got %d", cfg.Database.Port)
	}
	if cfg.Database.Name != "test_db" {
		t.Errorf("Expected name test_db, got %s", cfg.Database.Name)
	}

	// Verify JWT config
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("Expected secret test-secret, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.AccessExpiration != 30 {
		t.Errorf("Expected access expiration 30, got %d", cfg.JWT.AccessExpiration)
	}

	// Verify Redis config
	if cfg.Redis.Host != "testredis" {
		t.Errorf("Expected redis host testredis, got %s", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("Expected redis port 6380, got %d", cfg.Redis.Port)
	}

	// Verify Logger config
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Logger.Level)
	}
	if cfg.Logger.Path != "./test.log" {
		t.Errorf("Expected log path ./test.log, got %s", cfg.Logger.Path)
	}
}

func TestLoadConfig_JSON(t *testing.T) {
	// Create a temporary JSON config file
	content := `{
  "server": {
    "port": ":8888",
    "mode": "test"
  },
  "database": {
    "host": "jsondb",
    "port": 3308,
    "name": "json_db",
    "username": "jsonuser",
    "password": "jsonpass"
  },
  "jwt": {
    "secret": "json-secret",
    "access_expiration": 20,
    "refresh_expiration": 10
  },
  "redis": {
    "host": "jsonredis",
    "port": 6381,
    "password": "",
    "db": 2
  },
  "logger": {
    "level": "warn",
    "path": "./json.log",
    "max_size": 200,
    "max_age": 14,
    "max_backups": 5,
    "compress": true
  }
}`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify Server config
	if cfg.Server.Port != ":8888" {
		t.Errorf("Expected port :8888, got %s", cfg.Server.Port)
	}
	if cfg.Server.Mode != "test" {
		t.Errorf("Expected mode test, got %s", cfg.Server.Mode)
	}

	// Verify Database config
	if cfg.Database.Host != "jsondb" {
		t.Errorf("Expected host jsondb, got %s", cfg.Database.Host)
	}

	// Verify JWT config
	if cfg.JWT.Secret != "json-secret" {
		t.Errorf("Expected secret json-secret, got %s", cfg.JWT.Secret)
	}

	// Verify Logger config
	if cfg.Logger.Level != "warn" {
		t.Errorf("Expected log level warn, got %s", cfg.Logger.Level)
	}
	if cfg.Logger.Compress != true {
		t.Errorf("Expected compress true, got %v", cfg.Logger.Compress)
	}
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Create a minimal config file
	content := `
server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
  password: "file-password"
jwt:
  secret: "file-secret"
redis:
  host: "localhost"
  port: 6379
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Set environment variables (these should override file values)
	os.Setenv("KADMIN_SERVER_PORT", ":7777")
	os.Setenv("KADMIN_JWT_SECRET", "env-secret")
	os.Setenv("KADMIN_DATABASE_PASSWORD", "env-password")
	defer func() {
		os.Unsetenv("KADMIN_SERVER_PORT")
		os.Unsetenv("KADMIN_JWT_SECRET")
		os.Unsetenv("KADMIN_DATABASE_PASSWORD")
	}()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify environment variables override file values
	if cfg.Server.Port != ":7777" {
		t.Errorf("Expected port :7777 from env, got %s", cfg.Server.Port)
	}
	if cfg.JWT.Secret != "env-secret" {
		t.Errorf("Expected secret env-secret from env, got %s", cfg.JWT.Secret)
	}
	if cfg.Database.Password != "env-password" {
		t.Errorf("Expected password env-password from env, got %s (file had: file-password)", cfg.Database.Password)
	}
}

func TestValidateConfig_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing server port",
			config: Config{
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: true,
			errorMsg:    "server.port is required",
		},
		{
			name: "missing database host",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: true,
			errorMsg:    "database.host is required",
		},
		{
			name: "missing database port",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Host: "localhost", Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: true,
			errorMsg:    "database.port is required",
		},
		{
			name: "missing jwt secret",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: true,
			errorMsg:    "jwt.secret is required",
		},
		{
			name: "missing redis host",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Port: 6379},
			},
			expectError: true,
			errorMsg:    "redis.host is required",
		},
		{
			name: "invalid server mode",
			config: Config{
				Server:   ServerConfig{Port: ":8080", Mode: "invalid"},
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: true,
			errorMsg:    "server.mode must be one of: debug, release, test",
		},
		{
			name: "invalid logger level",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
				Logger:   LoggerConfig{Level: "invalid"},
			},
			expectError: true,
			errorMsg:    "logger.level must be one of: debug, info, warn, error, fatal",
		},
		{
			name: "valid config with defaults",
			config: Config{
				Server:   ServerConfig{Port: ":8080"},
				Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
				JWT:      JWTConfig{Secret: "secret"},
				Redis:    RedisConfig{Host: "localhost", Port: 6379},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateConfig_DefaultValues(t *testing.T) {
	config := Config{
		Server:   ServerConfig{Port: ":8080"},
		Database: DatabaseConfig{Host: "localhost", Port: 3306, Name: "test", Username: "root"},
		JWT:      JWTConfig{Secret: "secret"},
		Redis:    RedisConfig{Host: "localhost", Port: 6379},
	}

	err := validateConfig(&config)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Check default values are set
	if config.Server.Mode != "debug" {
		t.Errorf("Expected default mode 'debug', got %s", config.Server.Mode)
	}
	if config.Database.MaxIdleConns != 10 {
		t.Errorf("Expected default MaxIdleConns 10, got %d", config.Database.MaxIdleConns)
	}
	if config.Database.MaxOpenConns != 100 {
		t.Errorf("Expected default MaxOpenConns 100, got %d", config.Database.MaxOpenConns)
	}
	if config.JWT.AccessExpiration != 15 {
		t.Errorf("Expected default AccessExpiration 15, got %d", config.JWT.AccessExpiration)
	}
	if config.JWT.RefreshExpiration != 7 {
		t.Errorf("Expected default RefreshExpiration 7, got %d", config.JWT.RefreshExpiration)
	}
	if config.Logger.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logger.Level)
	}
	if config.Logger.Path != "./logs/app.log" {
		t.Errorf("Expected default log path './logs/app.log', got %s", config.Logger.Path)
	}
	if config.Logger.MaxSize != 100 {
		t.Errorf("Expected default MaxSize 100, got %d", config.Logger.MaxSize)
	}
	if config.Logger.MaxAge != 7 {
		t.Errorf("Expected default MaxAge 7, got %d", config.Logger.MaxAge)
	}
	if config.Logger.MaxBackups != 3 {
		t.Errorf("Expected default MaxBackups 3, got %d", config.Logger.MaxBackups)
	}
}
