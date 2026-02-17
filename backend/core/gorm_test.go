package core

import (
	"context"
	"testing"
	"time"

	"k-admin-system/config"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// TestInitDB_Success tests successful database initialization
func TestInitDB_Success(t *testing.T) {
	// Skip if no database available
	t.Skip("Requires MySQL database - run manually with test database")

	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         3306,
			Name:         "k_admin_test",
			Username:     "root",
			Password:     "",
			MaxIdleConns: 5,
			MaxOpenConns: 10,
		},
	}

	log, _ := zap.NewDevelopment()

	db, err := InitDB(cfg, log)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if db == nil {
		t.Fatal("Expected non-nil database instance")
	}

	// Verify connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL DB: %v", err)
	}

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != cfg.Database.MaxOpenConns {
		t.Errorf("Expected MaxOpenConns=%d, got %d", cfg.Database.MaxOpenConns, stats.MaxOpenConnections)
	}
}

// TestInitDB_InvalidHost tests connection failure with invalid host
func TestInitDB_InvalidHost(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Database: config.DatabaseConfig{
			Host:         "invalid-host-12345",
			Port:         3306,
			Name:         "test_db",
			Username:     "root",
			Password:     "",
			MaxIdleConns: 5,
			MaxOpenConns: 10,
		},
	}

	log, _ := zap.NewDevelopment()

	_, err := InitDB(cfg, log)
	if err == nil {
		t.Fatal("Expected error for invalid host, got nil")
	}
}

// TestGormLogger_LogMode tests log level setting
func TestGormLogger_LogMode(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "debug",
		},
	}

	gormLog := newGormLogger(log, cfg)

	// Test changing log mode
	newLogger := gormLog.LogMode(logger.Silent)
	if newLogger == nil {
		t.Fatal("Expected non-nil logger after LogMode")
	}

	// Verify it returns a new instance
	if newLogger == gormLog {
		t.Error("LogMode should return a new logger instance")
	}
}

// TestGormLogger_SlowQueryThreshold tests slow query detection
func TestGormLogger_SlowQueryThreshold(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "debug",
		},
	}

	gormLog := newGormLogger(log, cfg).(*gormLogger)

	// Verify default slow threshold
	expectedThreshold := 200 * time.Millisecond
	if gormLog.slowThreshold != expectedThreshold {
		t.Errorf("Expected slow threshold %v, got %v", expectedThreshold, gormLog.slowThreshold)
	}
}

// TestGormLogger_LogLevelByMode tests log level based on server mode
func TestGormLogger_LogLevelByMode(t *testing.T) {
	log, _ := zap.NewDevelopment()

	tests := []struct {
		mode     string
		expected logger.LogLevel
	}{
		{"debug", logger.Info},
		{"test", logger.Warn},
		{"release", logger.Error},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					Mode: tt.mode,
				},
			}

			gormLog := newGormLogger(log, cfg).(*gormLogger)
			if gormLog.logLevel != tt.expected {
				t.Errorf("Mode %s: expected log level %v, got %v", tt.mode, tt.expected, gormLog.logLevel)
			}
		})
	}
}

// TestGormLogger_InfoWarnError tests basic logging methods
func TestGormLogger_InfoWarnError(t *testing.T) {
	log, _ := zap.NewDevelopment()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "debug",
		},
	}

	gormLog := newGormLogger(log, cfg)

	// These should not panic
	ctx := context.TODO()
	gormLog.Info(ctx, "test info message")
	gormLog.Warn(ctx, "test warn message")
	gormLog.Error(ctx, "test error message")
}
